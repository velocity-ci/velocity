module Component.BuildTimeline exposing (..)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Time.DateTime as DateTime exposing (DateTime, DateTimeDelta)
import String exposing (padRight, split, join)


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


view : State -> Html msg
view points =
    div [ class "timeline" ]
        [ div [ class "line" ]
            [ div [ class "events" ] (viewPoints points)
            ]
        , viewDuration points
        ]


viewDuration : State -> Html msg
viewDuration state =
    let
        { hours, minutes, seconds } =
            duration state
    in
        small [ class "pull-right" ]
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


viewPoints : State -> List (Html msg)
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


viewPoint : Float -> Float -> Point -> Html msg
viewPoint ratio start point =
    let
        timestamp =
            DateTime.toTimestamp point.dateTime

        pos =
            eventPosition ratio start timestamp
    in
        div [ class "event", style [ "left" => (pos ++ "%") ] ]
            [ div [ class "circle" ]
                [ div [ class "circle-inner" ] []
                , div [ class "label" ]
                    [ label [] [ text point.label ]
                    , time [] [ text (formatTimeSeconds point.dateTime) ]
                    ]
                ]
            ]


eventPosition : Float -> Float -> Float -> String
eventPosition ratio start point =
    toFixed 2 ((point - start) * ratio)



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
        (round (value * power) |> toFloat)
            / power
            |> toString
            |> String.split "."
            |> pad
            |> String.join "."
