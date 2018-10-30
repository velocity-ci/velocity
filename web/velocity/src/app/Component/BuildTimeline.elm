module Component.BuildTimeline exposing (Config, Point, PopoverUpdate(..), State, StepPopovers, TimelinePopovers, addCompletedPoint, addCreatedPoint, addStepPoints, duration, endTime, eventPosition, mapPoints, pluralizeOrDrop, pointAndNext, pointContainerStyle, pointPercentage, pointStyle, points, ratio, shouldShowDuration, startTime, stepToPoint, toFixed, updatePopovers, view, viewDuration, viewPoint, viewPoints, viewTimeline)

-- EXTERNAL --
-- INTERNAL --

import Array exposing (Array)
import Bootstrap.Popover as Popover
import Css exposing (..)
import Data.Build as Build exposing (Build)
import Data.BuildOutput as BuildOutput exposing (Step)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.Task as ProjectTask exposing (Task)
import Dict exposing (Dict)
import Html exposing (Html)
import Html.Styled exposing (..)
import Html.Styled.Attributes as Attributes exposing (class, classList, css)
import Html.Styled.Events exposing (onClick)
import Page.Helpers exposing (formatTimeSeconds)
import String exposing (join, padRight, split)
import Time.DateTime as DateTime exposing (DateTime, DateTimeDelta)
import Util exposing ((=>))
import Views.Build


-- MODEL --


type alias State =
    List Point


type alias TimelinePopovers =
    { queued : Popover.State
    , completed : Popover.State
    , steps : StepPopovers
    }


type alias Config msg =
    { popoverMsg : PopoverUpdate -> Popover.State -> msg
    , clickPopoverMsg : PopoverUpdate -> msg
    }


type alias StepPopovers =
    Dict String Popover.State


type alias Point =
    { label : Maybe String
    , dateTime : DateTime
    , status : BuildStep.Status
    , popover : Popover.State
    , updateType : PopoverUpdate
    }


type PopoverUpdate
    = Queued
    | Step BuildStep.Id
    | Completed


points : Build -> TimelinePopovers -> ProjectTask.Task -> State
points build popovers task =
    []
        |> addCreatedPoint build popovers.queued
        |> addStepPoints build task popovers.steps
        |> addCompletedPoint build popovers.completed
        |> List.sortWith (\a b -> DateTime.compare a.dateTime b.dateTime)


addCreatedPoint : Build -> Popover.State -> List Point -> State
addCreatedPoint { createdAt } popover points =
    { label = Just "Queued"
    , dateTime = createdAt
    , status = BuildStep.Waiting
    , popover = popover
    , updateType = Queued
    }
        :: points


addCompletedPoint : Build -> Popover.State -> List Point -> State
addCompletedPoint { completedAt, status } popover points =
    let
        stepStatus =
            case status of
                Build.Running ->
                    BuildStep.Running

                Build.Failed ->
                    BuildStep.Failed

                Build.Success ->
                    BuildStep.Success

                Build.Waiting ->
                    BuildStep.Waiting

        label =
            Build.statusToString status
                |> Util.capitalize

        durationSecs =
            points
                |> duration
                |> .seconds

        lastPointDateTime =
            points
                |> List.reverse
                |> List.head
                |> Maybe.map .dateTime
                |> Maybe.map (DateTime.addSeconds (durationSecs // 2))
                |> Maybe.withDefault DateTime.epoch
    in
        case completedAt of
            Just dateTime ->
                { label = Just label
                , dateTime = dateTime
                , status = stepStatus
                , popover = popover
                , updateType = Completed
                }
                    :: points

            Nothing ->
                { label = Nothing
                , dateTime = lastPointDateTime
                , status = stepStatus
                , popover = popover
                , updateType = Completed
                }
                    :: points


addStepPoints : Build -> ProjectTask.Task -> StepPopovers -> List Point -> State
addStepPoints build task popovers points =
    build
        |> .steps
        |> List.filterMap (BuildOutput.joinSteps task)
        |> List.filterMap (stepToPoint popovers)
        |> List.append points


stepToPoint : StepPopovers -> Step -> Maybe Point
stepToPoint popovers ( taskStep, buildStep ) =
    Maybe.map2
        (\popover dateTime ->
            { label = Just (ProjectTask.stepName taskStep)
            , dateTime = dateTime
            , status = buildStep.status
            , popover = popover
            , updateType = Step buildStep.id
            }
        )
        (Dict.get (BuildStep.idToString buildStep.id) popovers)
        buildStep.startedAt


startTime : State -> Float
startTime points =
    points
        |> List.head
        |> Maybe.map (.dateTime >> DateTime.toTimestamp)
        |> Maybe.withDefault (DateTime.toTimestamp DateTime.epoch)


endTime : State -> Float
endTime points =
    points
        |> List.reverse
        |> startTime


duration : State -> DateTimeDelta
duration state =
    let
        start =
            DateTime.fromTimestamp (startTime state)

        end =
            DateTime.fromTimestamp (endTime state)
    in
        DateTime.delta end start


ratio : Float -> Float -> Float
ratio startTime endTime =
    100 / (endTime - startTime)



-- UPDATE --


updatePopovers : PopoverUpdate -> Popover.State -> TimelinePopovers -> TimelinePopovers
updatePopovers update state popovers =
    case update of
        Queued ->
            { popovers | queued = state }

        Completed ->
            { popovers | completed = state }

        Step stepId ->
            { popovers | steps = Dict.insert (BuildStep.idToString stepId) state popovers.steps }



-- VIEW --


view : Config msg -> State -> Html.Html msg
view config points =
    div
        [ css
            [ position relative ]
        , class "build-timeline"
        ]
        [ viewTimeline config points
        , if shouldShowDuration points then
            viewDuration points
          else
            text ""
        ]
        |> toUnstyled


viewTimeline : Config msg -> State -> Html.Styled.Html msg
viewTimeline config points =
    div
        [ css
            [ position relative
            , padding3 (Css.em 1.5) (Css.em 0) (Css.em 0.5)
            , borderWidth2 (px 0) (px 1)
            ]
        ]
        [ div
            [ css
                [ display block
                , width (pct 100)
                , position relative
                ]
            ]
            [ div
                [ css
                    [ position relative
                    , top (Css.em -0.8)
                    , margin2 (Css.em 0) (Css.em 0.5)
                    ]
                ]
                (points
                    |> mapPoints
                    |> viewPoints config
                )
            ]
        ]


shouldShowDuration : State -> Bool
shouldShowDuration points =
    case List.head (List.reverse points) of
        Just point ->
            List.member point.status [ BuildStep.Success, BuildStep.Failed ]

        Nothing ->
            False


viewDuration : State -> Html.Styled.Html msg
viewDuration state =
    let
        { hours, minutes, seconds } =
            duration state
    in
        Html.Styled.small
            [ css
                [ position absolute
                , right (px 0)
                , top (px -20)
                ]
            , class "py-2"
            ]
            [ i [ class "fa fa-clock-o" ] []
            , span []
                [ text " Ran for "
                , text (pluralizeOrDrop "hour" hours)
                , text (pluralizeOrDrop "min" (remainderBy 60 minutes))
                , text (pluralizeOrDrop "sec" (remainderBy 60 seconds))
                ]
            ]


pluralizeOrDrop : String -> Int -> String
pluralizeOrDrop word amount =
    case amount of
        0 ->
            ""

        1 ->
            toString amount ++ " " ++ word ++ " "

        _ ->
            toString amount ++ " " ++ word ++ "s "


mapPoints : State -> List ( Point, Maybe Point )
mapPoints points =
    points
        |> List.length
        |> List.range 1
        |> List.filterMap (pointAndNext points)


viewPoints : Config msg -> List ( Point, Maybe Point ) -> List (Html.Styled.Html msg)
viewPoints config combinedPoints =
    let
        points =
            List.map Tuple.first combinedPoints

        start =
            startTime points

        end =
            endTime points

        lineRatio =
            ratio start end
    in
        List.map
            (viewPoint config (ratio start end) start)
            combinedPoints


pointAndNext : List Point -> Int -> Maybe ( Point, Maybe Point )
pointAndNext listPoints index =
    let
        points =
            Array.fromList listPoints

        maybePoint =
            Array.get (index - 1) points

        maybeNext =
            Array.get index points
    in
        Maybe.map (\p -> ( p, maybeNext )) maybePoint


viewPoint : Config msg -> Float -> Float -> ( Point, Maybe Point ) -> Html.Styled.Html msg
viewPoint { popoverMsg, clickPopoverMsg } ratio start ( point, nextPoint ) =
    let
        popoverAttrs =
            popoverMsg point.updateType
                |> Popover.onHover point.popover
                |> List.map Attributes.fromUnstyled

        pointCircle =
            div
                (List.concat [ popoverAttrs, [ onClick (clickPopoverMsg point.updateType) ] ])
                [ div
                    [ class borderClass
                    , css [ pointStyle ]
                    ]
                    []
                ]

        popoverDirection =
            case point.updateType of
                Queued ->
                    Popover.right

                Step _ ->
                    Popover.bottom

                Completed ->
                    Popover.left

        pointLeft =
            pointPercentage ratio start point

        nextPointLeft =
            nextPoint
                |> Maybe.map (pointPercentage ratio start)
                |> Maybe.withDefault pointLeft

        nextRemainder =
            nextPointLeft - pointLeft

        borderCss =
            case point.status of
                BuildStep.Waiting ->
                    [ borderTopStyle dashed ]

                BuildStep.Running ->
                    [ borderTopStyle dotted ]

                BuildStep.Success ->
                    []

                BuildStep.Failed ->
                    []

        borderClass =
            case point.status of
                BuildStep.Waiting ->
                    "border-secondary"

                BuildStep.Running ->
                    "border-primary slide-in"

                BuildStep.Success ->
                    "border-success"

                BuildStep.Failed ->
                    "border-danger"

        popover =
            case point.label of
                Just label ->
                    Popover.config (toUnstyled pointCircle)
                        |> popoverDirection
                        |> Popover.titleH6 [] [ Html.text label ]
                        |> Popover.content [] [ Html.text (formatTimeSeconds point.dateTime) ]
                        |> Popover.view point.popover
                        |> fromUnstyled

                Nothing ->
                    text ""
    in
        div []
            [ div
                [ class borderClass
                , css
                    [ margin3 (Css.em 0.5) (Css.em 0) (Css.em 0)
                    , padding (Css.em 0)
                    , borderTop2 (px 4) solid
                    , width (pct nextRemainder)
                    , position absolute
                    , left (pct pointLeft)
                    , Css.batch borderCss
                    ]
                ]
                []
            , div
                [ css
                    [ pointContainerStyle
                    , left (pct pointLeft)
                    ]
                ]
                [ popover
                ]
            ]


pointPercentage : Float -> Float -> Point -> Float
pointPercentage ratio start { dateTime } =
    dateTime
        |> DateTime.toTimestamp
        |> eventPosition ratio start


pointStyle : Style
pointStyle =
    Css.batch
        [ position relative
        , backgroundColor (hex "ffffff")
        , border2 (px 4) solid
        , width (Css.em 1.3)
        , height (Css.em 1.3)
        , borderRadius (pct 50)
        ]


pointContainerStyle : Style
pointContainerStyle =
    Css.batch
        [ position absolute
        , display block
        , top (Css.em 0)
        , margin4 (Css.em 0) (Css.em 0) (Css.em 0) (Css.em -0.6)
        , cursor pointer
        ]


eventPosition : Float -> Float -> Float -> Float
eventPosition ratio start point =
    (point - start) * ratio



-- UTIL --


toFixed : Int -> Float -> String
toFixed precision value =
    let
        power =
            toFloat 10 ^ toFloat precision

        pad num =
            case num of
                [ x, y ] ->
                    [ x, String.padRight precision '0' y ]

                [ val ] ->
                    [ val, String.padRight precision '0' "" ]

                val ->
                    val
    in
        (Basics.round (value * power) |> toFloat)
            / power
            |> toString
            |> String.split "."
            |> pad
            |> String.join "."
