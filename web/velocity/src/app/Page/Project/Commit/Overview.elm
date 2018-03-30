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
import Views.Build exposing (viewBuildStatusIcon, viewBuildTextClass)


-- MODEL --


type alias Model =
    {}


initialModel : Model
initialModel =
    {}



-- VIEW --


view : Project -> Commit -> List ProjectTask.Task -> List Build -> Html Msg
view project commit tasks builds =
    div []
        [ viewTaskList project commit tasks builds
        ]


viewTaskList : Project -> Commit -> List ProjectTask.Task -> List Build -> Html Msg
viewTaskList project commit tasks builds =
    let
        taskList =
            List.map (viewTaskListItem project commit builds) tasks
                |> div [ class "list-group list-group-flush" ]
    in
        div [ class "card mt-3" ]
            [ h5 [ class "card-header" ] [ text "Tasks" ]
            , taskList
            ]


taskBuilds : ProjectTask.Task -> List Build -> List Build
taskBuilds task builds =
    builds
        |> List.filter (\b -> ProjectTask.idEquals task.id b.task.id)


maybeBuildFromTask : ProjectTask.Task -> List Build -> Maybe Build
maybeBuildFromTask task builds =
    taskBuilds task builds
        |> List.head


viewTaskListItem : Project -> Commit -> List Build -> ProjectTask.Task -> Html Msg
viewTaskListItem project commit builds task =
    let
        buildNum =
            List.length (taskBuilds task builds)

        routeTabParam =
            if buildNum > 0 then
                Just ("build-" ++ (toString buildNum))
            else
                Nothing

        route =
            CommitRoute.Task task.name routeTabParam
                |> ProjectRoute.Commit commit.hash
                |> Route.Project project.slug

        maybeBuild =
            maybeBuildFromTask task builds

        textClass =
            maybeBuild
                |> Maybe.map viewBuildTextClass
                |> Maybe.withDefault ""
    in
        a
            [ class (textClass ++ " list-group-item list-group-item-action flex-column align-items-center justify-content-between")
            , Route.href route
            , onClickPage NewUrl route
            ]
            [ div [ class "" ] [ h5 [ class "mb-1" ] [ text (ProjectTask.nameToString task.name) ] ]
            , p [ class "mb-1" ] [ text task.description ]
            ]



-- UPDATE --


type Msg
    = NewUrl String


update : Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        NewUrl url ->
            model => Navigation.newUrl url
