module Page.Project.Task exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask exposing (Step(..), BuildStep, RunStep)
import Request.Project
import Http
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Views.Page as Page
import Task exposing (Task)
import Util exposing ((=>))
import Page.Project.Route as ProjectRoute
import Route exposing (Route)
import Page.Project.Commit as Commit


-- MODEL --


type alias Model =
    { commit : Commit
    , task : ProjectTask.Task
    }


init : Session -> Project.Id -> Commit.Hash -> ProjectTask.Name -> Task PageLoadError Model
init session id hash name =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommit =
            maybeAuthToken
                |> Request.Project.commit id hash
                |> Http.toTask

        loadTask =
            maybeAuthToken
                |> Request.Project.commitTask id hash name
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map2 Model loadCommit loadTask
            |> Task.mapError handleLoadError



-- VIEW --


view : Model -> Html Msg
view model =
    viewStepList model.task.steps
        |> div []


viewStepList : List Step -> List (Html Msg)
viewStepList steps =
    let
        stepView i step =
            let
                stepNum =
                    i + 1
            in
                case step of
                    Run runStep ->
                        viewRunStep stepNum runStep

                    Build buildStep ->
                        viewBuildStep stepNum buildStep
    in
        List.indexedMap stepView steps


viewBuildStep : Int -> BuildStep -> Html Msg
viewBuildStep i buildStep =
    let
        tagList =
            List.map (\t -> li [] [ text t ]) buildStep.tags
                |> ul []

        rightDl =
            dl []
                [ dt [] [ text "Tags" ]
                , dd [] [ tagList ]
                ]

        leftDl =
            dl []
                [ dt [] [ text "Context" ]
                , dd [] [ text buildStep.context ]
                , dt [] [ text "Dockerfile" ]
                , dd [] [ text buildStep.dockerfile ]
                ]
    in
        div [ class "card default-margin-bottom" ]
            [ div [ class "card-header" ]
                [ h5 [ class "mb-0" ]
                    [ text (toString i ++ ". " ++ buildStep.description) ]
                ]
            , div [ class "card-body" ]
                [ div [ class "row" ]
                    [ div [ class "col-md-6" ] [ leftDl ]
                    , div [ class "col-md-6" ] [ rightDl ]
                    ]
                ]
            ]


viewRunStep : Int -> RunStep -> Html Msg
viewRunStep i runStep =
    let
        command =
            String.join " " runStep.command

        envTable =
            table [ class "table" ]
                [ tbody []
                    (List.map
                        (\( k, v ) ->
                            tr []
                                [ th [] [ text k ]
                                , td [] [ text v ]
                                ]
                        )
                        runStep.environment
                    )
                ]

        ignoreExitCode =
            runStep.ignoreExitCode
                |> toString
                |> String.toLower

        leftDl =
            dl []
                [ dt [] [ text "Ignore exit code" ]
                , dd [] [ text ignoreExitCode ]
                , dt [] [ text "Image" ]
                , dd [] [ text runStep.image ]
                , dt [] [ text "Mount point" ]
                , dd [] [ text runStep.mountPoint ]
                , dt [] [ text "Working dir" ]
                , dd [] [ text runStep.workingDir ]
                , dt [] [ text "Command" ]
                , dd [] [ text command ]
                ]
    in
        div [ class "card default-margin-bottom" ]
            [ div [ class "card-header" ]
                [ h5 [ class "mb-0" ]
                    [ text (toString i ++ ". " ++ runStep.description) ]
                ]
            , div [ class "card-body" ]
                [ div [ class "row" ]
                    [ div [ class "col-md-6" ] [ leftDl ]
                    , div [ class "col-md-6" ] [ envTable ]
                    ]
                ]
            ]


breadcrumb : Project -> Commit -> ProjectTask.Task -> List ( Route, String )
breadcrumb project commit task =
    List.concat
        [ Commit.breadcrumb project commit
        , [ ( Route.Project (ProjectRoute.Task commit.hash task.name) project.id, ProjectTask.nameToString task.name ) ]
        ]



-- UPDATE --


type Msg
    = NoOp


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    model => Cmd.none
