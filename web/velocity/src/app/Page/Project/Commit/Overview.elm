module Page.Project.Commit.Overview exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask
import Data.Build as Build exposing (Build)
import Page.Helpers exposing (formatDateTime)
import Navigation
import Util exposing ((=>))
import Route
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Helpers exposing (onClickPage)


-- MODEL --


type alias Model =
    {}


initialModel : Model
initialModel =
    {}



-- VIEW --


view : Project -> Commit -> List ProjectTask.Task -> Html Msg
view project commit tasks =
    div []
        [ div [ class "card mt-3" ] [ div [ class "card-body" ] [ viewCommitDetails commit ] ]
        , viewTaskList project commit tasks
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
            , dt [] [ i [ attribute "aria-hidden" "true", class "fa fa-file-code-o" ] [] ]
            , dd [] [ text hash ]
            , dt [] [ text "Author" ]
            , dd [] [ text commit.author ]
            , dt [] [ i [ attribute "aria-hidden" "true", class "fa fa-calendar" ] [] ]
            , dd [] [ text (formatDateTime commit.date) ]
            ]


viewTaskList : Project -> Commit -> List ProjectTask.Task -> Html Msg
viewTaskList project commit tasks =
    let
        taskList =
            List.map (viewTaskListItem project commit) tasks
                |> div [ class "list-group list-group-flush" ]
    in
        div [ class "card default-margin-top" ]
            [ h5 [ class "card-header" ] [ text "Tasks" ]
            , taskList
            ]


viewTaskListItem : Project -> Commit -> ProjectTask.Task -> Html Msg
viewTaskListItem project commit task =
    let
        route =
            CommitRoute.Task task.name
                |> ProjectRoute.Commit commit.hash
                |> Route.Project project.id
    in
        a [ class "list-group-item list-group-item-action flex-column align-items-center", Route.href route, onClickPage NewUrl route ]
            [ div [ class "d-flex w-100 justify-content-between" ]
                [ h5 [ class "mb-1" ] [ text (ProjectTask.nameToString task.name) ]
                ]
            , p [ class "mb-1" ] [ text task.description ]
            ]



-- UPDATE --


type Msg
    = NewUrl String


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        NewUrl url ->
            model => Navigation.newUrl url
