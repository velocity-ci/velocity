module Page.Project.Commit exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (formatDateTime, sortByDatetime)
import Page.Project.Commits as Commits
import Request.Project
import Util exposing ((=>))
import Task exposing (Task)
import Views.Page as Page
import Http
import Route exposing (Route)
import Page.Project.Route as ProjectRoute


-- MODEL --


type alias Model =
    { commit : Commit
    , tasks : List ProjectTask.Task
    }


init : Session -> Project.Id -> Commit.Hash -> Task PageLoadError Model
init session id hash =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommit =
            maybeAuthToken
                |> Request.Project.commit id hash
                |> Http.toTask

        loadTasks =
            maybeAuthToken
                |> Request.Project.commitTasks id hash
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map2 Model loadCommit loadTasks
            |> Task.mapError handleLoadError



-- VIEW --


view : Model -> Html Msg
view model =
    let
        commit =
            model.commit
    in
        div []
            [ div [ class "card" ] [ div [ class "card-body" ] [ viewCommitDetails commit ] ]
            , viewTaskList model.tasks
            ]


viewCommitDetails : Commit -> Html Msg
viewCommitDetails commit =
    let
        hash =
            Commit.hashToString commit.hash
    in
        dl [ style [ ( "margin-bottom", "0" ) ] ]
            [ dt [] [ text "Message" ]
            , dd [] [ text commit.message ]
            , dt [] [ text "Commit" ]
            , dd [] [ text hash ]
            , dt [] [ text "Author" ]
            , dd [] [ text commit.author ]
            , dt [] [ text "Date" ]
            , dd [] [ text (formatDateTime commit.date) ]
            ]


viewTaskList : List ProjectTask.Task -> Html Msg
viewTaskList tasks =
    let
        taskList =
            List.map viewTaskListItem tasks
                |> div [ class "list-group list-group-flush" ]
    in
        div [ class "card first-row" ]
            [ h5 [ class "card-header" ] [ text "Tasks" ]
            , taskList
            ]


viewTaskListItem : ProjectTask.Task -> Html Msg
viewTaskListItem task =
    a [ class "list-group-item list-group-item-action flex-column align-items-start", href "#" ]
        [ div [ class "d-flex w-100 justify-content-between" ] [ h5 [ class "mb-1" ] [ text task.name ] ]
        , p [ class "mb-1" ] [ text task.description ]
        ]


breadcrumb : Project -> Commit -> List ( Route, String )
breadcrumb project commit =
    List.concat
        [ Commits.breadcrumb project
        , [ ( Route.Project (ProjectRoute.Commit commit.hash) project.id, Commit.truncateHash commit.hash ) ]
        ]



-- UPDATE --


type Msg
    = NoOp


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    model => Cmd.none
