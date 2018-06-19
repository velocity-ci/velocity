module Component.BuildLog exposing (..)

{- A stateful BuildOutput component.
   I plan to convert this to a stateless component soon.
-}
-- INTERNAL

import Context exposing (Context)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (Id, BuildStream, BuildStreamOutput)
import Data.AuthToken as AuthToken exposing (AuthToken)
import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Request.Build
import Request.Errors
import Util exposing ((=>))
import Page.Helpers exposing (formatDateTime, formatTimeSeconds)
import Views.Build exposing (..)
import Ports


-- EXTERNAL

import Html exposing (..)
import Html.Attributes exposing (..)
import Array exposing (Array)
import Dict exposing (Dict)
import Task exposing (Task)
import Time.DateTime as DateTime exposing (DateTime)
import Ansi.Log
import Json.Encode as Encode
import Json.Decode as Decode
import Dom.Scroll as Scroll
import Html.Lazy exposing (lazy)


-- MODEL


type alias Model =
    { log : Log
    , autoScrollMessages : Bool
    }


type alias TaskStep =
    ProjectTask.Step


type alias Step =
    ( TaskStep, BuildStep )


type alias BuildStepId =
    String


type alias BuildStreamId =
    String


type alias BuildStreamIndex =
    Int


type alias LineNumber =
    Int


type alias Log =
    Dict BuildStepId LogStep


type alias LogStep =
    { step : Step
    , streams : Dict BuildStreamId LogStepStream
    }


type alias LogStepStream =
    { buildStream : BuildStream
    , lines : Dict LineNumber LogStepStreamLine
    }


type alias LogStepStreamLine =
    { updates : List BuildStreamOutput
    , ansi : Ansi.Log.Model
    }


init : Context -> ProjectTask.Task -> Maybe AuthToken -> Build -> Task Request.Errors.HttpError Model
init context task maybeAuthToken build =
    build
        |> loadBuildStreams context task maybeAuthToken
        |> Task.map initialModel


initialModel : Log -> Model
initialModel log =
    { log = log
    , autoScrollMessages = True
    }


loadBuildStreams : Context -> ProjectTask.Task -> Maybe AuthToken -> Build -> Task Request.Errors.HttpError Log
loadBuildStreams context task maybeAuthToken build =
    build
        |> .steps
        |> List.sortBy .number
        |> List.filterMap (joinSteps task)
        |> List.map (resolveLogStepStream context maybeAuthToken)
        |> List.foldl (++) []
        |> Task.sequence
        |> Task.map (List.foldl insertStream Dict.empty)


insertStream : ( LogStepStream, Step ) -> Dict String LogStep -> Dict String LogStep
insertStream ( { buildStream, lines }, step ) dict =
    let
        ( _, buildStep ) =
            step

        outputStream =
            { buildStream = buildStream
            , lines = lines
            }

        logStepDictKey =
            logStepKey buildStep

        streamDictKey =
            streamKey buildStream

        insert =
            case Dict.get logStepDictKey dict of
                Just exists ->
                    { exists | streams = Dict.insert streamDictKey outputStream exists.streams }

                Nothing ->
                    { step = step, streams = Dict.singleton streamDictKey outputStream }
    in
        Dict.insert logStepDictKey insert dict


linesArrayToDict : Array BuildStreamOutput -> Dict LineNumber LogStepStreamLine
linesArrayToDict lines =
    Array.foldl (\v a -> Dict.insert v.line (outputToLogLine v) a) Dict.empty lines


outputToLogLine : BuildStreamOutput -> LogStepStreamLine
outputToLogLine buildStreamOutput =
    let
        ansi =
            Ansi.Log.init Ansi.Log.Cooked
                |> Ansi.Log.update buildStreamOutput.output
    in
        { updates = [ buildStreamOutput ]
        , ansi = ansi
        }


joinSteps : ProjectTask.Task -> BuildStep -> Maybe Step
joinSteps task buildStep =
    task
        |> .steps
        |> Array.fromList
        |> Array.get buildStep.number
        |> Maybe.map (\taskStep -> ( taskStep, buildStep ))


resolveLogStepStream : Context -> Maybe AuthToken -> Step -> List (Task Request.Errors.HttpError ( LogStepStream, Step ))
resolveLogStepStream context maybeAuthToken step =
    let
        ( _, buildStep ) =
            step
    in
        buildStep
            |> .streams
            |> List.map
                (\buildStream ->
                    buildStream.id
                        |> Request.Build.streamOutput context maybeAuthToken
                        |> Task.map (\lines -> ( { buildStream = buildStream, lines = linesArrayToDict lines }, step ))
                )



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions _ =
    scrolledToBottom


scrolledToBottom : Sub Msg
scrolledToBottom =
    Decode.decodeValue Decode.bool
        >> Result.toMaybe
        >> Maybe.withDefault False
        |> Ports.onScrolledToBottom
        |> Sub.map ScrolledToBottom



-- CHANNELS --


streamChannelName : BuildStream -> String
streamChannelName stream =
    "stream:" ++ (BuildStream.idToString stream.id)


events : Model -> Dict String (List ( String, Encode.Value -> Msg ))
events model =
    let
        foldStreamEvents ( buildStepId, streams ) dict =
            streams
                |> List.foldl
                    (\stream acc ->
                        let
                            events =
                                [ ( "streamLine:new", AddStreamOutput buildStepId stream ) ]
                        in
                            Dict.insert (streamChannelName stream) events acc
                    )
                    dict
    in
        model.log
            |> Dict.foldl (\buildStepId val acc -> ( buildStepId, List.map .buildStream (Dict.values val.streams) ) :: acc) []
            |> List.foldl foldStreamEvents Dict.empty


leaveChannels : Model -> List String
leaveChannels model =
    Dict.keys (events model)



-- UPDATE --


type Msg
    = AddStreamOutput BuildStepId BuildStream Encode.Value
    | ScrolledToBottom Bool
    | NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        AddStreamOutput buildStepId buildStream outputJson ->
            let
                log =
                    case decodeBuildStreamOutput outputJson of
                        Just buildStreamOutput ->
                            updateLog ( buildStepId, streamKey buildStream, buildStreamOutput ) model.log

                        Nothing ->
                            model.log

                scrollCmd =
                    if model.autoScrollMessages then
                        Task.attempt (always NoOp) (Scroll.toBottom "scroll-id")
                    else
                        Cmd.none
            in
                { model | log = log }
                    => scrollCmd

        ScrolledToBottom isScrolled ->
            { model | autoScrollMessages = isScrolled }
                => Cmd.none

        NoOp ->
            model => Cmd.none


decodeBuildStreamOutput : Encode.Value -> Maybe BuildStreamOutput
decodeBuildStreamOutput outputValue =
    outputValue
        |> Decode.decodeValue BuildStream.outputDecoder
        |> Result.toMaybe


updateLog : ( BuildStepId, BuildStreamId, BuildStreamOutput ) -> Log -> Log
updateLog ( logStepKey, buildStreamId, buildStreamOutput ) log =
    Dict.update logStepKey (Maybe.map <| updateLogStep ( buildStreamId, buildStreamOutput )) log


updateLogStep : ( BuildStreamId, BuildStreamOutput ) -> LogStep -> LogStep
updateLogStep ( streamDictKey, buildStreamOutput ) logStep =
    let
        streams =
            logStep
                |> .streams
                |> Dict.update streamDictKey (Maybe.map <| updateLogStepStream buildStreamOutput)
    in
        { logStep | streams = streams }


updateLogStepStream : BuildStreamOutput -> LogStepStream -> LogStepStream
updateLogStepStream buildStreamOutput logStepStream =
    let
        line =
            updateLogStepStreamLine buildStreamOutput logStepStream
    in
        { logStepStream | lines = Dict.insert buildStreamOutput.line line logStepStream.lines }


updateLogStepStreamLine : BuildStreamOutput -> LogStepStream -> LogStepStreamLine
updateLogStepStreamLine buildStreamOutput logStepStream =
    case Dict.get buildStreamOutput.line logStepStream.lines of
        Just existingLine ->
            { existingLine
                | updates = buildStreamOutput :: existingLine.updates
                , ansi = Ansi.Log.update buildStreamOutput.output existingLine.ansi
            }

        Nothing ->
            outputToLogLine buildStreamOutput



-- VIEW --


type alias ViewStepLine =
    { ansi : Ansi.Log.Model
    , lineNumber : LineNumber
    , streamIndex : BuildStreamIndex
    , streamId : BuildStreamId
    , streamName : String
    , timestamp : DateTime
    }


view : Build -> Model -> Html Msg
view build { log } =
    let
        ansiOutput =
            log
                |> Dict.toList
                |> List.sortBy (\( _, { step } ) -> step |> Tuple.second |> .number)
                |> List.map (viewStepContainer build)
    in
        div [] ansiOutput


viewStepContainer : Build -> ( BuildStepId, LogStep ) -> Html Msg
viewStepContainer build ( stepId, logStep ) =
    let
        ( taskStep, buildStep ) =
            logStep.step

        buildStep_ =
            build.steps
                |> List.filter (\s -> s.id == buildStep.id)
                |> List.head
    in
        case buildStep_ of
            Just step ->
                div
                    [ class "mt-3 build-info-container border-secondary"
                    , classList (buildStepBorderColourClassList step)
                    ]
                    [ h5
                        [ class "d-flex justify-content-between"
                        , classList (headerBackgroundColourClassList step)
                        ]
                        [ text (viewCardTitle taskStep)
                        , text " "
                        , viewBuildStepStatusIcon step
                        ]
                    , div [ class "p-0 small" ] [ lazy viewStepLog logStep ]
                    ]

            Nothing ->
                text ""


viewStepLog : LogStep -> Html Msg
viewStepLog step =
    div [ class "table-responsive" ] [ viewStepLogTable step ]


viewStepLogTable : LogStep -> Html Msg
viewStepLogTable step =
    step
        |> flattenStepLines
        |> List.map viewLine
        |> table [ class "table table-sm table-hover mb-0" ]


viewLine : ViewStepLine -> Html Msg
viewLine { timestamp, streamName, ansi, streamIndex } =
    tr [ class "d-flex" ]
        [ td [ class "col-1" ] [ span [ classList [ "badge" => True, streamBadgeClass streamIndex => True ] ] [ text streamName ] ]
        , td [ class "col-1" ] [ span [ class "badge badge-light" ] [ text (formatTimeSeconds timestamp) ] ]
        , td [ class "col-10" ] [ viewLineAnsi ansi.lines ]
        ]


viewLineAnsi : Array Ansi.Log.Line -> Html Msg
viewLineAnsi lines =
    lines
        |> Array.get ((Array.length lines) - 2)
        |> Maybe.map Ansi.Log.viewLine
        |> Maybe.withDefault (text "")



-- HELPERS --


flattenStepLines : LogStep -> List ViewStepLine
flattenStepLines logStep =
    logStep
        |> .streams
        |> Dict.toList
        |> List.indexedMap mapStream
        |> List.foldl (++) []
        |> List.sortWith sortLines


mapStream : BuildStreamIndex -> ( BuildStreamId, LogStepStream ) -> List ViewStepLine
mapStream streamIndex ( streamId, stream ) =
    stream
        |> .lines
        |> Dict.values
        |> List.filterMap
            (\line ->
                line.updates
                    |> List.head
                    |> Maybe.map
                        (\lastUpdate ->
                            { ansi = line.ansi
                            , lineNumber = lastUpdate.line
                            , streamIndex = streamIndex
                            , streamId = streamId
                            , streamName = stream.buildStream.name
                            , timestamp = lastUpdate.timestamp
                            }
                        )
            )


sortLines : ViewStepLine -> ViewStepLine -> Order
sortLines a b =
    if a.streamId == b.streamId then
        Basics.compare a.lineNumber b.lineNumber
    else
        DateTime.compare a.timestamp b.timestamp


logStepKey : BuildStep -> BuildStepId
logStepKey buildStep =
    BuildStep.idToString buildStep.id


streamKey : BuildStream -> BuildStreamId
streamKey buildStream =
    BuildStream.idToString buildStream.id
