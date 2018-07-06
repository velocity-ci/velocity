module Component.BuildTimeline exposing (..)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Time.DateTime as DateTime exposing (DateTime)
import String exposing (padRight, split, join)


-- INTERNAL --

import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.Task as ProjectTask exposing (Task)
import Data.BuildOutput as BuildOutput exposing (Step)
import Util exposing ((=>))


-- MODEL --


type alias State =
    List Point


type alias Point =
    ( Float, Step )


points : Task -> List BuildStep -> State
points task steps =
    steps
        |> List.filterMap (BuildOutput.joinSteps task)
        |> List.filterMap point


point : Step -> Maybe Point
point ( taskStep, buildStep ) =
    Maybe.map DateTime.toTimestamp buildStep.startedAt
        |> Maybe.map (\timestamp -> ( timestamp, ( taskStep, buildStep ) ))



--points : State -> List Float
--points steps =
--    [ 0, 5, 10, 20, 100 ]


startTime : State -> Float
startTime points =
    points
        |> List.head
        |> Maybe.map Tuple.first
        |> Maybe.withDefault (DateTime.toTimestamp DateTime.epoch)


endTime : State -> Float
endTime points =
    points
        |> List.reverse
        |> startTime


ratio : Float -> Float -> Float
ratio startTime endTime =
    100 / (endTime - startTime)



-- VIEW --


line : State -> Html msg
line points =
    div [ class "timeline" ]
        [ div [ class "line" ]
            [ div [ class "events" ] (events points)
            ]
        ]


events : State -> List (Html msg)
events points =
    let
        start =
            startTime points

        end =
            endTime points

        lineRatio =
            ratio start end
    in
        List.map (event (ratio start end) start) <|
            points


event : Float -> Float -> Point -> Html msg
event ratio start ( point, _ ) =
    let
        pos =
            eventPosition ratio start point
    in
        div [ class "event", style [ "left" => (pos ++ "%") ] ]
            [ div [ class "circle" ]
                [ div [ class "circle-inner" ] []
                , div [ class "label" ]
                    [ label [] [ text pos ]
                    , time [] [ text pos ]
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
