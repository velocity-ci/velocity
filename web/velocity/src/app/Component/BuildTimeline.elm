module Component.BuildTimeline exposing (..)

-- EXTERNAL --

import Time.DateTime as DateTime exposing (DateTime, DateTimeDelta)
import String exposing (padRight, split, join)
import Html exposing (Html)
import Html.Styled.Attributes as Attributes exposing (css, class, classList)
import Html.Styled exposing (..)
import Css exposing (..)
import Bootstrap.Popover as Popover
import Dict exposing (Dict)


-- INTERNAL --

import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.Task as ProjectTask exposing (Task)
import Data.BuildOutput as BuildOutput exposing (Step)
import Util exposing ((=>))
import Page.Helpers exposing (formatTimeSeconds)


-- MODEL --


type alias State =
    List Point


type PointColour
    = Green


type alias TimelinePopovers =
    { queued : Popover.State
    , completed : Popover.State
    , steps : StepPopovers
    }


type alias Config msg =
    { popoverMsg : PopoverUpdate -> Popover.State -> msg }


type alias StepPopovers =
    Dict String Popover.State


type alias Point =
    { label : String
    , dateTime : DateTime
    , color : PointColour
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
    { label = "Queued"
    , dateTime = createdAt
    , color = Green
    , popover = popover
    , updateType = Queued
    }
        :: points


addCompletedPoint : Build -> Popover.State -> List Point -> State
addCompletedPoint { completedAt } popover points =
    case completedAt of
        Just dateTime ->
            { label = "Completed"
            , dateTime = dateTime
            , color = Green
            , popover = popover
            , updateType = Completed
            }
                :: points

        Nothing ->
            points


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
            { label = ProjectTask.stepName taskStep
            , dateTime = dateTime
            , color = Green
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
    points
        |> viewTimeline config
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
            [ class "border-success"
            , css
                [ display block
                , listStyleType none
                , margin3 (Css.em 0.5) (Css.em 0) (Css.em 0)
                , padding (Css.em 0)
                , height (Css.em 1)
                , borderTop2 (px 4) solid
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
                (viewPoints config points)
            ]
        , viewDuration points
        ]


viewDuration : State -> Html.Styled.Html msg
viewDuration state =
    let
        { hours, minutes, seconds } =
            duration state
    in
        Html.Styled.small [ css [ float right ] ]
            [ i [ class "fa fa-clock-o" ] []
            , text " Ran for "
            , text (pluralizeOrDrop "hour" hours)
            , text (pluralizeOrDrop "min" minutes)
            , text (pluralizeOrDrop "sec" seconds)
            ]


pluralizeOrDrop : String -> Int -> String
pluralizeOrDrop word amount =
    case amount of
        0 ->
            ""

        1 ->
            (toString amount) ++ " " ++ word ++ " "

        _ ->
            (toString amount) ++ " " ++ word ++ "s "


viewPoints : Config msg -> State -> List (Html.Styled.Html msg)
viewPoints config points =
    let
        start =
            startTime points

        end =
            endTime points

        lineRatio =
            ratio start end
    in
        List.map (viewPoint config (ratio start end) start) <|
            points


viewPoint : Config msg -> Float -> Float -> Point -> Html.Styled.Html msg
viewPoint { popoverMsg } ratio start point =
    let
        popoverAttrs =
            popoverMsg point.updateType
                |> Popover.onHover point.popover
                |> List.map Attributes.fromUnstyled

        pointCircle =
            div
                popoverAttrs
                [ div
                    [ class "border-success"
                    , css
                        [ position relative
                        , backgroundColor (hex "ffffff")
                        , border2 (px 4) solid
                        , width (Css.em 1.3)
                        , height (Css.em 1.3)
                        , borderRadius (pct 50)
                        ]
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
    in
        div
            [ css
                [ left (pct <| eventPosition ratio start <| DateTime.toTimestamp point.dateTime)
                , width (Css.em 1)
                , height (Css.em 1)
                , position absolute
                , top (Css.em 0)
                , margin4 (Css.em 0) (Css.em 0) (Css.em 0) (Css.em -0.6)
                ]
            ]
            [ Popover.config (toUnstyled pointCircle)
                |> popoverDirection
                |> Popover.titleH4 [] [ Html.text point.label ]
                |> Popover.content [] [ Html.text (formatTimeSeconds point.dateTime) ]
                |> Popover.view point.popover
                |> fromUnstyled
            ]



--viewPointPopover : Point -> Html.Styled.Html msg
--viewPointPopover point =
--    Pop


eventPosition : Float -> Float -> Float -> Float
eventPosition ratio start point =
    ((point - start) * ratio)



-- UTIL --


toFixed : Int -> Float -> String
toFixed precision value =
    let
        power =
            toFloat 10 ^ (toFloat precision)

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
