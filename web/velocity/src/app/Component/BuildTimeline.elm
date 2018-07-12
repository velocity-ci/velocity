module Component.BuildTimeline exposing (..)

-- EXTERNAL --

import Time.DateTime as DateTime exposing (DateTime, DateTimeDelta)
import String exposing (padRight, split, join)
import Html exposing (Html)
import Html.Styled.Attributes exposing (css, class, classList)
import Html.Styled exposing (..)
import Css exposing (..)


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


type alias Point =
    { label : String
    , dateTime : DateTime
    , color : PointColour
    }


points : Build -> ProjectTask.Task -> State
points build task =
    []
        |> addCreatedPoint build
        |> addStepPoints build task
        |> addCompletedPoint build
        |> List.sortWith (\a b -> DateTime.compare a.dateTime b.dateTime)


addCreatedPoint : Build -> List Point -> State
addCreatedPoint { createdAt } points =
    { label = "Queued"
    , dateTime = createdAt
    , color = Green
    }
        :: points


addCompletedPoint : Build -> List Point -> State
addCompletedPoint { completedAt } points =
    case completedAt of
        Just dateTime ->
            { label = "Completed"
            , dateTime = dateTime
            , color = Green
            }
                :: points

        Nothing ->
            points


addStepPoints : Build -> ProjectTask.Task -> List Point -> State
addStepPoints build task points =
    build
        |> .steps
        |> List.filterMap (BuildOutput.joinSteps task)
        |> List.filterMap stepToPoint
        |> List.append points


stepToPoint : Step -> Maybe Point
stepToPoint ( taskStep, buildStep ) =
    buildStep
        |> .startedAt
        |> Maybe.map
            (\dateTime ->
                { label = ProjectTask.stepName taskStep
                , dateTime = dateTime
                , color = Green
                }
            )


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



-- VIEW --


view : State -> Html.Html msg
view points =
    points
        |> viewTimeline
        |> toUnstyled


viewTimeline : State -> Html.Styled.Html msg
viewTimeline points =
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
                , listStyleType none
                , margin3 (Css.em 0.5) (Css.em 0) (Css.em 0)
                , padding (Css.em 0)
                , height (Css.em 1)
                , borderTop3 (px 4) solid (hex "008000")
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
                (viewPoints points)
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


viewPoints : State -> List (Html.Styled.Html msg)
viewPoints points =
    let
        start =
            startTime points

        end =
            endTime points

        lineRatio =
            ratio start end
    in
        List.map (viewPoint (ratio start end) start) <|
            points


viewPoint : Float -> Float -> Point -> Html.Styled.Html msg
viewPoint ratio start point =
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
        [ div
            [ css
                [ position relative
                , backgroundColor (hex "ffffff")
                , border3 (px 4) solid (hex "008000")
                , width (Css.em 1.3)
                , height (Css.em 1.3)
                , borderRadius (pct 50)
                ]
            ]
            [ div
                [ css
                    [ display block
                    , width (Css.em 10)
                    , padding2 (Css.em 0.5) (Css.em 1)
                    , textAlign center
                    , position absolute
                    , margin4 (px -2) (px 0) (px 0) (Css.em -5)
                    , left (pct 50)
                    , top (pct 50)
                    ]
                ]
                [ label
                    [ css
                        [ display block
                        , fontWeight bold
                        , margin4 (px 0) (px 0) (px 5) (px 0)
                        ]
                    ]
                    [ text point.label ]
                , time [] [ text (formatTimeSeconds point.dateTime) ]
                ]
            ]
        ]


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
