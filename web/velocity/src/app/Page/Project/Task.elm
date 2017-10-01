module Page.Project.Task exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
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
    , toggledStep : Maybe Step
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

        initialModel commit task =
            let
                toggledStep =
                    task.steps
                        |> List.head
            in
                { commit = commit
                , task = task
                , toggledStep = toggledStep
                }
    in
        Task.map2 initialModel loadCommit loadTask
            |> Task.mapError handleLoadError



-- VIEW --


view : Model -> Html Msg
view model =
    viewStepList model.task.steps model.toggledStep
        |> div []


viewStepList : List Step -> Maybe Step -> List (Html Msg)
viewStepList steps toggledStep =
    let
        stepView i step =
            let
                stepNum =
                    i + 1

                runStep =
                    viewRunStep stepNum

                buildStep =
                    viewBuildStep stepNum
            in
                case ( step, toggledStep ) of
                    ( Run run, Just (Run toggled) ) ->
                        runStep run (run == toggled)

                    ( Build build, Just (Build toggled) ) ->
                        buildStep build (build == toggled)

                    ( Run run, _ ) ->
                        runStep run False

                    ( Build build, _ ) ->
                        buildStep build False
    in
        List.indexedMap stepView steps


viewBuildStep : Int -> BuildStep -> Bool -> Html Msg
viewBuildStep i buildStep toggled =
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

        title =
            toString i ++ ". " ++ buildStep.description
    in
        viewStepCollapse (Build buildStep) title toggled <|
            [ div [ class "row" ]
                [ div [ class "col-md-6" ] [ leftDl ]
                , div [ class "col-md-6" ] [ rightDl ]
                ]
            ]


viewRunStep : Int -> RunStep -> Bool -> Html Msg
viewRunStep i runStep toggled =
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

        title =
            toString i ++ ". " ++ runStep.description
    in
        viewStepCollapse (Run runStep) title toggled <|
            [ div [ class "row" ]
                [ div [ class "col-md-6" ] [ leftDl ]
                , div [ class "col-md-6" ] [ envTable ]
                ]
            ]


viewStepCollapse : Step -> String -> Bool -> List (Html Msg) -> Html Msg
viewStepCollapse step title toggled contents =
    let
        msg =
            if toggled then
                ToggleStep Nothing
            else
                ToggleStep (Just step)

        caretClassList =
            [ ( "fa-caret-square-o-down", toggled )
            , ( "fa-caret-square-o-up", not toggled )
            ]
    in
        div [ class "card" ]
            [ div [ class "card-header d-flex justify-content-between align-items-center", onClick msg ]
                [ h5 [ class "mb-0" ] [ text title ]
                , button
                    [ type_ "button"
                    , class "btn"
                    ]
                    [ i [ class "fa", classList caretClassList ] []
                    ]
                ]
            , div
                [ class "collapse"
                , classList [ ( "show", toggled ) ]
                ]
                [ div [ class "card-body" ] contents
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
    = ToggleStep (Maybe Step)


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        ToggleStep maybeStep ->
            { model | toggledStep = maybeStep }
                => Cmd.none
