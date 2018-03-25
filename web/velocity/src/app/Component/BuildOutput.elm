module Component.BuildOutput exposing (Model, init, view)

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
import Page.Helpers exposing (validClasses, formatDateTime)
import Views.Build exposing (viewBuildStatusIcon, viewBuildStepStatusIcon, viewBuildTextClass)


-- EXTERNAL

import Html exposing (..)
import Html.Attributes exposing (..)
import Array exposing (Array)
import Dict exposing (Dict)
import Task exposing (Task)
import Ansi.Log


-- MODEL


type alias Model =
    { build : Build
    , outputStreams : OutputStreams
    }


type alias OutputStream =
    { buildStepNumber : Int
    , taskStep : ProjectTask.Step
    , buildStepId : BuildStep.Id
    , ansi : Ansi.Log.Model
    }


type alias OutputStreams =
    Dict String OutputStream


init :
    Context
    -> ProjectTask.Task
    -> Maybe AuthToken
    -> Build
    -> Task Request.Errors.HttpError Model
init context task maybeAuthToken build =
    let
        initialModel outputStreams =
            { build = build
            , outputStreams = outputStreams
            }
    in
        build
            |> loadBuildStreams context task maybeAuthToken
            |> Task.map initialModel


loadBuildStreams :
    Context
    -> ProjectTask.Task
    -> Maybe AuthToken
    -> Build
    -> Task Request.Errors.HttpError OutputStreams
loadBuildStreams context task maybeAuthToken build =
    build.steps
        |> List.sortBy .number
        |> List.map
            (\buildStep ->
                let
                    maybeTaskStep =
                        task.steps
                            |> Array.fromList
                            |> Array.get buildStep.number
                in
                    ( maybeTaskStep, buildStep )
            )
        |> List.map
            (\( maybeTaskStep, buildStep ) ->
                List.map
                    (\buildStream ->
                        Request.Build.streamOutput context maybeAuthToken buildStream.id
                            |> Task.map (\output -> ( buildStream, maybeTaskStep, buildStep, output ))
                    )
                    buildStep.streams
            )
        |> List.foldr (++) []
        |> Task.sequence
        |> Task.map
            (List.foldr
                (\( buildStream, maybeTaskStep, buildStep, outputStreams ) dict ->
                    case maybeTaskStep of
                        Just taskStep ->
                            let
                                ansiInit =
                                    Ansi.Log.init Ansi.Log.Cooked

                                lineAnsi outputLine ansi =
                                    Ansi.Log.update outputLine.output ansi

                                ansi =
                                    Array.foldl lineAnsi ansiInit outputStreams

                                outputStream =
                                    { buildStepNumber = buildStep.number
                                    , taskStep = taskStep
                                    , buildStepId = buildStep.id
                                    , ansi = ansi
                                    }
                            in
                                Dict.insert (BuildStream.idToString buildStream.id) outputStream dict

                        Nothing ->
                            dict
                )
                Dict.empty
            )



-- VIEW


view : Model -> Html msg
view { build, outputStreams } =
    let
        ansiOutput =
            outputStreams
                |> Dict.toList
                |> List.sortBy (\( _, outputStream ) -> outputStream.buildStepNumber)
                |> List.map
                    (\( streamId, { taskStep, buildStepId, ansi } ) ->
                        let
                            ansiView =
                                Ansi.Log.view ansi

                            buildStep =
                                build.steps
                                    |> List.filter (\s -> s.id == buildStepId)
                                    |> List.head
                        in
                            case buildStep of
                                Just buildStep ->
                                    if buildStep.status == BuildStep.Waiting then
                                        text ""
                                    else
                                        div
                                            [ class "card mt-3"
                                            , classList (buildStepBorderColourClassList buildStep)
                                            ]
                                            [ h5
                                                [ class "card-header d-flex justify-content-between"
                                                , classList (headerBackgroundColourClassList buildStep)
                                                ]
                                                [ text (viewCardTitle taskStep)
                                                , text " "
                                                , viewBuildStepStatusIcon buildStep
                                                ]
                                            , div [ class "card-body text-white" ] [ ansiView ]
                                            ]

                                _ ->
                                    text ""
                    )
    in
        div [] (viewBuildInformation build :: ansiOutput)


viewBuildInformation : Build -> Html msg
viewBuildInformation build =
    div [ class "card mt-3", classList (buildCardClassList build) ]
        [ div [ class "card-body" ]
            [ dl [ class "row mb-0" ]
                [ dt [ class "col-sm-3" ] [ text "Id" ]
                , dd [ class "col-sm-9" ] [ text (Build.idToString build.id) ]
                , dt [ class "col-sm-3" ] [ text "Created" ]
                , dd [ class "col-sm-9" ] [ text (formatDateTime build.createdAt) ]
                , dt [ class "col-sm-3" ] [ text "Started" ]
                , dd [ class "col-sm-9" ] [ text (Maybe.map formatDateTime build.startedAt |> Maybe.withDefault "-") ]
                , dt [ class "col-sm-3" ] [ text "Completed" ]
                , dd [ class "col-sm-9" ] [ text (Maybe.map formatDateTime build.completedAt |> Maybe.withDefault "-") ]
                , dt [ class "col-sm-3" ] [ text "Status" ]
                , dd [ class "col-sm-9" ] [ text (Build.statusToString build.status) ]
                ]
            ]
        ]


headerBackgroundColourClassList : BuildStep -> List ( String, Bool )
headerBackgroundColourClassList buildStep =
    case buildStep.status of
        BuildStep.Waiting ->
            []

        BuildStep.Running ->
            []

        BuildStep.Success ->
            [ "bg-success" => True ]

        BuildStep.Failed ->
            [ "bg-danger" => True ]


buildStepBorderColourClassList : BuildStep -> List ( String, Bool )
buildStepBorderColourClassList buildStep =
    case buildStep.status of
        BuildStep.Waiting ->
            [ "border" => True
            , "border-light" => True
            ]

        BuildStep.Running ->
            [ "border" => True
            , "border-primary" => True
            ]

        BuildStep.Success ->
            [ "border" => True
            , "border-success" => True
            , "text-white" => True
            ]

        BuildStep.Failed ->
            [ "border" => True
            , "border-danger" => True
            , "text-white" => True
            ]


buildCardClassList : Build -> List ( String, Bool )
buildCardClassList build =
    case build.status of
        Build.Success ->
            [ "border-success" => True
            , "text-success" => True
            ]

        Build.Failed ->
            [ "border-danger" => True
            , "text-danger" => True
            ]

        _ ->
            []


viewCardTitle : ProjectTask.Step -> String
viewCardTitle taskStep =
    case taskStep of
        Build _ ->
            "Build"

        Run _ ->
            "Run"

        Clone _ ->
            "Clone"

        Compose _ ->
            "Compose"

        Push _ ->
            "Push"
