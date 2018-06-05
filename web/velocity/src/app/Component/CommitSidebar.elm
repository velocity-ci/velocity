module Component.CommitSidebar exposing (view, Context, Config)

-- INTERNAL

import Data.Commit as Commit exposing (Commit)
import Data.Task as Task exposing (Task)
import Data.Project as Project exposing (Project)
import Data.Build as Build exposing (Build)
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Commit exposing (branchList, infoPanel)
import Views.Helpers exposing (onClickPage)
import Views.Build exposing (viewBuildStatusIconClasses, viewBuildTextClass)
import Util exposing ((=>))


-- EXTERNAL

import Html exposing (..)
import Html.Attributes exposing (..)


-- CONFIG


type alias Config msg =
    { newUrlMsg : String -> msg }


type alias Context =
    { project : Project
    , builds : List Build
    , commit : Commit
    , tasks : List Task
    , selected : Maybe Task
    }


type alias NavTaskProperties =
    { isSelected : Bool
    , route : Route
    , iconClass : String
    , textClass : String
    , itemText : String
    }



-- VIEW


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


{-| List of task navigation
-}
taskNav : Config msg -> Context -> Html msg
taskNav config context =
    let
        tasks =
            context
                |> .tasks
                |> filterTasks
                |> sortTasks

        taskItem =
            taskNavItem config.newUrlMsg

        toProperties =
            taskNavProperties context
    in
        ul [ class "nav nav-pills flex-column project-navigation p-0" ] <|
            List.map (toProperties >> taskItem) tasks


{-| Single nav item for a task
-}
taskNavItem : (String -> msg) -> NavTaskProperties -> Html msg
taskNavItem newUrlMsg { isSelected, route, itemText, textClass } =
    li [ class "nav-item" ]
        [ a
            [ class "nav-link"
            , class textClass
            , classList [ "active" => isSelected ]
            , Route.href route
            , onClickPage newUrlMsg route
            ]
            [ text itemText
            ]
        ]



-- HELPERS


taskNavProperties : Context -> Task -> NavTaskProperties
taskNavProperties context task =
    { isSelected = isSelected context.selected task
    , route = taskToRoute context task
    , iconClass = taskIconClass context task
    , textClass = taskTextClass context task
    , itemText = Task.nameToString task.name
    }


{-| Filter out any tasks which have a blank name (this shouldn't be needed in the future)
-}
filterTasks : List Task -> List Task
filterTasks tasks =
    List.filter (.name >> Task.nameToString >> String.isEmpty >> not) tasks


{-| Sort tasks by name
-}
sortTasks : List Task -> List Task
sortTasks tasks =
    List.sortBy (.name >> Task.nameToString) tasks


{-| Filter builds by task
-}
taskBuilds : Task -> List Build -> List Build
taskBuilds task builds =
    List.filter (.task >> .id >> Task.idEquals task.id) builds


{-| Icon for a task based on its latest build
-}
taskIconClass : Context -> Task -> String
taskIconClass context task =
    task
        |> latestTaskBuild context
        |> Maybe.map viewBuildStatusIconClasses
        |> Maybe.withDefault ""


taskTextClass : Context -> Task -> String
taskTextClass context task =
    task
        |> latestTaskBuild context
        |> Maybe.map viewBuildTextClass
        |> Maybe.withDefault ""


{-| Get latest build for a task
-}
latestTaskBuild : Context -> Task -> Maybe Build
latestTaskBuild { builds } task =
    builds
        |> taskBuilds task
        |> List.head


{-| Determine if a task is currently selected
-}
isSelected : Maybe Task -> Task -> Bool
isSelected maybeTask task =
    case maybeTask of
        Just selected ->
            selected.id == task.id

        Nothing ->
            False


taskToRoute : Context -> (Task -> Route)
taskToRoute { project, commit } =
    taskRoute project commit


taskRoute : Project -> Commit -> Task -> Route
taskRoute project commit task =
    CommitRoute.Task task.name Nothing
        |> ProjectRoute.Commit commit.hash
        |> Route.Project project.slug
