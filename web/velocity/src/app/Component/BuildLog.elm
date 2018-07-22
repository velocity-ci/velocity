module Component.BuildLog exposing (..)

{- A stateful BuildOutput component.
   I plan to convert this to a stateless component soon.
-}
-- INTERNAL

import Context exposing (Context)
import Component.BuildTimeline
import Data.Build as Build exposing (Build)
import Data.BuildOutput as BuildOutput exposing (..)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (Id, BuildStream, BuildStreamOutput)
import Data.AuthToken as AuthToken exposing (AuthToken)
import Data.Task as ProjectTask
import Request.Build
import Request.Errors
import Util exposing ((=>))
import Page.Helpers exposing (formatDateTime, formatTimeSeconds)
import Views.Build exposing (..)
import Views.Task
import Ports
import Component.BuildTimeline as BuildTimeline


-- EXTERNAL

import Html exposing (Html)
import Html.Attributes
import Html.Events
import Html.Styled.Attributes as StyledAttributes exposing (css, class, classList)
import Html.Styled as Styled exposing (..)
import Html.Styled.Events exposing (onClick)
import Css exposing (..)
import Array exposing (Array)
import Dict exposing (Dict)
import Task exposing (Task)
import Time.DateTime as DateTime exposing (DateTime)
import Ansi.Log
import Json.Encode as Encode
import Json.Decode as Decode
import Dom.Scroll as Scroll
import Html.Styled.Lazy exposing (lazy)
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Button as Button
import Bootstrap.Modal as Modal
import Bootstrap.Popover as Popover
import Dom


-- MODEL


type alias Model =
    { log : Log
    , stepInfoModal : StepInfoModal
    , autoScrollMessages : Bool
    , timelinePopovers : BuildTimeline.TimelinePopovers
    }


type alias StepInfoModal =
    ( Modal.Visibility, Maybe ProjectTask.Step )


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
    , filterDropdown : Dropdown.State
    , collapsed : Bool
    }


type alias LogStepStream =
    { buildStream : BuildStream
    , lines : Dict LineNumber LogStepStreamLine
    , visible : Bool
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
    , stepInfoModal = ( Modal.hidden, Nothing )
    , autoScrollMessages = True
    , timelinePopovers = timelinePopovers log
    }


timelinePopovers : Log -> BuildTimeline.TimelinePopovers
timelinePopovers log =
    { queued = Popover.initialState
    , completed = Popover.initialState
    , steps = Dict.foldl (\id step acc -> Dict.insert id Popover.initialState acc) Dict.empty log
    }


timelineConfig : BuildTimeline.Config Msg
timelineConfig =
    { popoverMsg = UpdatePopover
    , clickPopoverMsg = ClickPopover
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
            , visible = True
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
                    { step = step
                    , streams = Dict.singleton streamDictKey outputStream
                    , filterDropdown = Dropdown.initialState
                    , collapsed = False
                    }
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
                        |> Task.map (\lines -> ( logStepStream buildStream lines, step ))
                )


logStepStream : BuildStream -> Array BuildStreamOutput -> LogStepStream
logStepStream buildStream lines =
    { buildStream = buildStream
    , lines = linesArrayToDict lines
    , visible = True
    }



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ scrolledToBottom
        , filterSubscriptions model.log
        , modalSubscriptions (Tuple.first model.stepInfoModal)
        ]


modalSubscriptions : Modal.Visibility -> Sub Msg
modalSubscriptions visibility =
    Modal.subscriptions visibility AnimateStepInfoModal


scrolledToBottom : Sub Msg
scrolledToBottom =
    Decode.decodeValue Decode.bool
        >> Result.toMaybe
        >> Maybe.withDefault False
        |> Ports.onScrolledToBottom
        |> Sub.map ScrolledToBottom


filterSubscriptions : Log -> Sub Msg
filterSubscriptions log =
    log
        |> Dict.values
        |> List.map logStepSubscriptions
        |> Sub.batch


logStepSubscriptions : LogStep -> Sub Msg
logStepSubscriptions logStep =
    let
        ( _, buildStep ) =
            logStep.step
    in
        buildStep.id
            |> ToggleStepFilterDropdown
            |> Dropdown.subscriptions logStep.filterDropdown



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
    | ToggleStepFilterDropdown BuildStep.Id Dropdown.State
    | ToggleStepCollapse BuildStep.Id
    | ToggleStreamVisibility BuildStep.Id BuildStream.Id
    | CloseStepInfoModal
    | AnimateStepInfoModal Modal.Visibility
    | OpenStepInfoModal LogStep
    | UpdatePopover BuildTimeline.PopoverUpdate Popover.State
    | ClickPopover BuildTimeline.PopoverUpdate
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

                scrollCmds =
                    if model.autoScrollMessages then
                        scrollToBottom
                    else
                        []
            in
                { model | log = log }
                    ! scrollCmds

        ScrolledToBottom isScrolled ->
            { model | autoScrollMessages = isScrolled }
                => Cmd.none

        ToggleStepFilterDropdown buildStepId state ->
            { model | log = updateLogStepDropdown buildStepId state model.log }
                => Cmd.none

        ToggleStreamVisibility buildStepId buildStreamId ->
            { model | log = toggleLogVisibility buildStepId buildStreamId model.log }
                => Cmd.none

        ToggleStepCollapse buildStepId ->
            { model | log = toggleStepCollapse buildStepId model.log }
                => Cmd.none

        CloseStepInfoModal ->
            { model | stepInfoModal = ( Modal.hidden, Nothing ) }
                => Cmd.none

        AnimateStepInfoModal visibility ->
            let
                ( _, maybeStep ) =
                    model.stepInfoModal
            in
                { model | stepInfoModal = ( visibility, maybeStep ) }
                    => Cmd.none

        OpenStepInfoModal logStep ->
            let
                ( taskStep, _ ) =
                    logStep.step
            in
                { model | stepInfoModal = ( Modal.shown, Just taskStep ) }
                    => Cmd.none

        UpdatePopover popover state ->
            { model | timelinePopovers = BuildTimeline.updatePopovers popover state model.timelinePopovers }
                => Cmd.none

        ClickPopover popover ->
            model
                ! (case popover of
                    BuildTimeline.Queued ->
                        []

                    BuildTimeline.Step stepId ->
                        [ Ports.scrollIntoView (BuildStep.idToString stepId) ]

                    BuildTimeline.Completed ->
                        scrollToBottom
                  )

        NoOp ->
            model => Cmd.none


scrollToBottom : List (Cmd Msg)
scrollToBottom =
    [ scrollTo Scroll.toBottom "scroll-html-id"
    , scrollTo Scroll.toBottom "scroll-body-id"
    ]


scrollTo : (Dom.Id -> Task Dom.Error ()) -> Dom.Id -> Cmd Msg
scrollTo scrollFn id =
    Task.attempt (always NoOp) (scrollFn id)


decodeBuildStreamOutput : Encode.Value -> Maybe BuildStreamOutput
decodeBuildStreamOutput outputValue =
    outputValue
        |> Decode.decodeValue BuildStream.outputDecoder
        |> Result.toMaybe


updateLogStepDropdown : BuildStep.Id -> Dropdown.State -> Log -> Log
updateLogStepDropdown buildStepId state log =
    let
        targetKey =
            BuildStep.idToString buildStepId

        updateDropdown currentKey step =
            if currentKey == targetKey then
                { step | filterDropdown = state }
            else
                { step | filterDropdown = Dropdown.initialState }
    in
        Dict.map updateDropdown log


toggleStepCollapse : BuildStep.Id -> Log -> Log
toggleStepCollapse buildStepId log =
    let
        dictKey =
            BuildStep.idToString buildStepId
    in
        Dict.update dictKey (Maybe.map (\step -> { step | collapsed = not step.collapsed })) log


toggleLogVisibility : BuildStep.Id -> BuildStream.Id -> Log -> Log
toggleLogVisibility buildStepId buildStreamId log =
    let
        dictKey =
            BuildStep.idToString buildStepId

        updateStep =
            Maybe.map (toggleLogStepVisibility buildStreamId)
    in
        Dict.update dictKey updateStep log


toggleLogStepVisibility : BuildStream.Id -> LogStep -> LogStep
toggleLogStepVisibility buildStreamId logStep =
    let
        dictKey =
            BuildStream.idToString buildStreamId

        updateStream =
            Maybe.map (\stream -> { stream | visible = not stream.visible })
    in
        { logStep | streams = Dict.update dictKey updateStream logStep.streams }


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


view : Build -> ProjectTask.Task -> Model -> Html.Html Msg
view build task model =
    let
        stepOutput =
            model.log
                |> Dict.toList
                |> List.sortBy (\( _, { step } ) -> step |> Tuple.second |> .number)
                |> List.map (viewStepContainer build)
    in
        div []
            [ div [] stepOutput
            , viewStepInfoModal model.stepInfoModal
            ]
            |> toUnstyled


viewTimeline : Build -> Model -> ProjectTask.Task -> Html.Html Msg
viewTimeline build { timelinePopovers } task =
    task
        |> BuildTimeline.points build timelinePopovers
        |> BuildTimeline.view timelineConfig


viewStepContainer : Build -> ( BuildStepId, LogStep ) -> Styled.Html Msg
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
                    [ StyledAttributes.fromUnstyled (Html.Attributes.id stepId)
                    , class "py-3"
                    , class (viewBuildStepBorderClass step)
                    , css
                        [ borderLeft2 (px 4) solid ]
                    ]
                    [ h5
                        [ class "px-2 d-flex justify-content-between"
                        , classList (headerBackgroundColourClassList step)
                        ]
                        [ text (ProjectTask.stepName taskStep)
                        , viewStepButtonToolbar buildStep_ logStep
                        ]
                    , div [ class "p-0 small" ]
                        [ lazy viewStepLog logStep ]
                    ]

            Nothing ->
                text ""


viewStepButtonToolbar : Maybe BuildStep -> LogStep -> Styled.Html Msg
viewStepButtonToolbar maybeBuildStep logStep =
    let
        showCollapseToggle =
            maybeBuildStep
                |> Maybe.map (\step -> step.status /= BuildStep.Waiting)
                |> Maybe.withDefault False

        showStreamFilter =
            (Dict.size logStep.streams) > 1 && not logStep.collapsed
    in
        div []
            [ Util.viewIfStyled showStreamFilter (viewStepStreamFilter logStep)
            , text " "
            , viewStepInfoButton logStep
            , text " "
            , Util.viewIfStyled showCollapseToggle (viewStepCollapseToggle logStep)
            ]


viewStepInfoButton : LogStep -> Styled.Html Msg
viewStepInfoButton logStep =
    Button.button
        [ Button.small
        , Button.light
        , Button.onClick (OpenStepInfoModal logStep)
        ]
        [ Html.text "Info" ]
        |> fromUnstyled


viewStepInfoModal : StepInfoModal -> Styled.Html Msg
viewStepInfoModal ( visibility, maybeStep ) =
    case maybeStep of
        Just taskStep ->
            Modal.config CloseStepInfoModal
                |> Modal.large
                |> Modal.withAnimation AnimateStepInfoModal
                |> Modal.hideOnBackdropClick True
                |> Modal.h3 [] [ Html.text "Modal header" ]
                |> Modal.body [] [ Views.Task.viewStepContents taskStep ]
                |> Modal.footer []
                    [ Button.button
                        [ Button.outlinePrimary
                        , Button.attrs [ Html.Events.onClick CloseStepInfoModal ]
                        ]
                        [ Html.text "Close" ]
                    ]
                |> Modal.view visibility
                |> fromUnstyled

        Nothing ->
            text ""


viewStepCollapseToggle : LogStep -> Styled.Html Msg
viewStepCollapseToggle logStep =
    let
        ( _, buildStep ) =
            logStep.step

        buttonText =
            if logStep.collapsed then
                "Show"
            else
                "Hide"
    in
        Button.button
            [ Button.small
            , Button.light
            , Button.onClick (ToggleStepCollapse buildStep.id)
            ]
            [ Html.text buttonText ]
            |> fromUnstyled


viewStepStreamFilter : LogStep -> Styled.Html Msg
viewStepStreamFilter logStep =
    let
        ( _, buildStep ) =
            logStep.step

        headerItem =
            Dropdown.header [ Html.text "Streams" ]

        streamItems =
            logStep.streams
                |> Dict.values
                |> List.indexedMap (,)
                |> List.sortBy (Tuple.second >> .buildStream >> .name)
                |> List.map (viewStepStreamFilterItem logStep)
    in
        Dropdown.dropdown
            logStep.filterDropdown
            { options =
                [ Dropdown.dropLeft
                , Dropdown.menuAttrs [ Html.Attributes.class "item-filter-dropdown" ]
                ]
            , toggleMsg = ToggleStepFilterDropdown buildStep.id
            , toggleButton =
                Dropdown.toggle [ Button.light, Button.small ] [ Html.text "Filter" ]
            , items = headerItem :: streamItems
            }
            |> fromUnstyled


viewStepStreamFilterItem : LogStep -> ( Int, LogStepStream ) -> Dropdown.DropdownItem Msg
viewStepStreamFilterItem logStep ( streamIndex, logStepStream ) =
    let
        ( _, buildStep ) =
            logStep.step

        msg =
            ToggleStreamVisibility buildStep.id logStepStream.buildStream.id
    in
        Dropdown.buttonItem
            [ onClickPreventDefault msg ]
            [ Html.i
                [ Html.Attributes.class "fa"
                , Html.Attributes.classList [ "fa-check" => logStepStream.visible ]
                ]
                []
            , Html.text " "
            , Html.span
                [ Html.Attributes.classList
                    [ "badge" => True
                    , streamBadgeClass streamIndex => True
                    ]
                ]
                [ Html.text logStepStream.buildStream.name ]
            ]


viewStepLog : LogStep -> Html.Html Msg
viewStepLog step =
    if step.collapsed then
        Html.text ""
    else
        div [ class "table-responsive" ] [ viewStepLogTable step ]
            |> toUnstyled


viewStepLogTable : LogStep -> Styled.Html Msg
viewStepLogTable step =
    step
        |> flattenStepLines
        |> List.map viewLine
        |> Styled.table [ class "table table-unbordered table-sm table-hover mb-0" ]


viewLine : ViewStepLine -> Styled.Html Msg
viewLine { timestamp, streamName, ansi, streamIndex } =
    tr [ class "d-flex" ]
        [ td [ class "d-none d-sm-none d-md-block col-1" ]
            [ span [ classList [ "badge" => True, streamBadgeClass streamIndex => True ] ] [ text streamName ] ]
        , td [ class "d-none d-sm-none d-md-block col-1" ]
            [ span [ class "badge badge-light" ] [ text (formatTimeSeconds timestamp) ] ]
        , td [ class "col-xs-12 col-sm-12 col-10" ]
            [ viewLineAnsi ansi.lines ]
        ]


viewLineAnsi : Array Ansi.Log.Line -> Styled.Html Msg
viewLineAnsi lines =
    lines
        |> Array.get ((Array.length lines) - 2)
        |> Maybe.map (Ansi.Log.viewLine >> fromUnstyled)
        |> Maybe.withDefault (text "")


viewBuildInformation : Build -> Styled.Html Msg
viewBuildInformation build =
    let
        dateText date =
            date
                |> Maybe.map formatDateTime
                |> Maybe.withDefault "-"
    in
        div [ class "card mt-3", classList (buildCardClassList build) ]
            [ div [ class "card-body" ]
                [ dl [ class "row mb-0" ]
                    [ dt [ class "col-sm-3" ] [ text "Created" ]
                    , dd [ class "col-sm-9" ] [ text (formatDateTime build.createdAt) ]
                    , dt [ class "col-sm-3" ] [ text "Started" ]
                    , dd [ class "col-sm-9" ] [ text (dateText build.startedAt) ]
                    , dt [ class "col-sm-3" ] [ text "Completed" ]
                    , dd [ class "col-sm-9" ] [ text (dateText build.completedAt) ]
                    , dt [ class "col-sm-3" ] [ text "Status" ]
                    , dd [ class "col-sm-9" ] [ text (Build.statusToString build.status) ]
                    ]
                ]
            ]



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
    if stream.visible then
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
    else
        []


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


onClickPreventDefault : msg -> Html.Attribute msg
onClickPreventDefault message =
    Html.Events.onWithOptions
        "click"
        { stopPropagation = True
        , preventDefault = False
        }
        (Decode.succeed message)
