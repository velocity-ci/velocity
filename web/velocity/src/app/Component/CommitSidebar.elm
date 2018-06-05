module Component.CommitSidebar exposing (view, Context, Config)

-- INTERNAL

import Data.Commit as Commit exposing (Commit)
import Data.Task as Task exposing (Task)
import Data.Project as Project exposing (Project)
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Commit exposing (branchList, infoPanel)
import Views.Helpers exposing (onClickPage)
import Util exposing ((=>))


-- EXTERNAL

import Html exposing (..)
import Html.Attributes exposing (..)


type alias Config msg =
    { newUrlMsg : String -> msg }


type alias Context =
    { project : Project
    , commit : Commit
    , tasks : List Task
    , selected : Maybe Task
    }


view : Config msg -> Context -> Html msg
view config context =
    div [ class "sub-sidebar" ]
        [ details context.commit
        , taskNav config context
        ]


details : Commit -> Html msg
details commit =
    div [ class "p-2" ]
        [ div [ class "card" ]
            [ div [ class "card-body" ]
                [ infoPanel commit
                , hr [] []
                , branchList commit
                ]
            ]
        ]


taskToRoute : Context -> (Task -> Route)
taskToRoute { project, commit } =
    taskRoute project commit


taskRoute : Project -> Commit -> Task -> Route
taskRoute project commit task =
    CommitRoute.Task task.name Nothing
        |> ProjectRoute.Commit commit.hash
        |> Route.Project project.slug


isSelected : Maybe Task -> Task -> Bool
isSelected maybeTask task =
    case maybeTask of
        Just selected ->
            selected.id == task.id

        Nothing ->
            False


taskNav : Config msg -> Context -> Html msg
taskNav config context =
    let
        tasks =
            context.tasks
                |> filterTasks
                |> sortTasks

        taskItem =
            context
                |> taskToRoute
                |> taskNavItem config.newUrlMsg (isSelected context.selected)
    in
        ul [ class "nav nav-pills flex-column project-navigation p-0" ] <|
            List.map taskItem tasks


filterTasks : List Task -> List Task
filterTasks tasks =
    List.filter (.name >> Task.nameToString >> String.isEmpty >> not) tasks


sortTasks : List Task -> List Task
sortTasks tasks =
    List.sortBy (.name >> Task.nameToString) tasks


taskNavItem : (String -> msg) -> (Task -> Bool) -> (Task -> Route) -> Task -> Html msg
taskNavItem newUrlMsg isSelected toRoute task =
    let
        route =
            toRoute task

        activeClassList =
            [ "active" => isSelected task ]
    in
        li [ class "nav-item" ]
            [ a
                [ class "nav-link"
                , classList activeClassList
                , Route.href route
                , onClickPage newUrlMsg route
                ]
                [ text <| Task.nameToString task.name ]
            ]
