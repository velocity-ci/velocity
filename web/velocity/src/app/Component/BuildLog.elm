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


type alias LineNumber =
    Int


type alias Log =
    Dict BuildStepId LogStep


type alias LogStep =
    { step : Step
    , streams : List LogStepStream
    }


type alias LogStepStream =
    { buildStream : BuildStream
    , lines : Dict LineNumber LogStepStreamLine
    }


type alias LogStepStreamLine =
    { updates : List BuildStreamOutput }


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

        key =
            logStepKey buildStep

        insert =
            case Dict.get key dict of
                Just exists ->
                    { exists | streams = exists.streams ++ [ outputStream ] }

                Nothing ->
                    { step = step, streams = [ outputStream ] }
    in
        Dict.insert key insert dict


logStepKey : BuildStep -> BuildStepId
logStepKey buildStep =
    BuildStep.idToString buildStep.id


linesArrayToDict : Array BuildStreamOutput -> Dict LineNumber LogStepStreamLine
linesArrayToDict lines =
    Array.foldl (\v a -> Dict.insert v.line (outputToLogLine v) a) Dict.empty lines


outputToLogLine : BuildStreamOutput -> LogStepStreamLine
outputToLogLine output =
    { updates = [ output ] }


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



-- UPDATE --


type Msg
    = AddStreamOutput BuildStepId BuildStream Encode.Value
    | ScrolledToBottom Bool
    | NoOp



--update : Msg -> Model -> ( Model, Cmd Msg )
--update msg model =
--    case msg of
--        AddStreamOutput buildStepId buildStream outputJson ->
--            let
--                outputStreams =
--                    addStreamOutput ( buildStepId, buildStream, outputJson ) model.outputStreams
--
--                scrollCmd =
--                    if model.autoScrollMessages then
--                        Task.attempt (always NoOp) (Scroll.toBottom "scroll-id")
--                    else
--                        Cmd.none
--            in
--                { model | outputStreams = outputStreams }
--                    => scrollCmd
--
--        ScrolledToBottom isScrolled ->
--            { model | autoScrollMessages = isScrolled }
--                => Cmd.none
--
--        NoOp ->
--            model => Cmd.none


updateLogStep : ( BuildStream.Id, BuildStreamOutput ) -> LogStep -> LogStep
updateLogStep ( buildStreamId, buildStreamOutput ) logStep =
    let
        streams =
            List.map
                (\stream ->
                    if stream.buildStream.id == buildStreamId then
                        updateLogStepStream buildStreamOutput stream
                    else
                        stream
                )
                logStep.streams
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
            { existingLine | updates = buildStreamOutput :: existingLine.updates }

        Nothing ->
            outputToLogLine buildStreamOutput
