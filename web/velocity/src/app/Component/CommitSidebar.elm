module Component.CommitSidebar exposing (view, Context, Config, displayType, DisplayType(..))

-- INTERNAL

import Data.Commit as Commit exposing (Commit)
import Data.Task as Task exposing (Task)
import Data.Project as Project exposing (Project)
import Data.Build as Build exposing (Build)
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Commit exposing (branchList, infoPanel, truncateCommitMessage)
import Views.Helpers exposing (onClickPage)
import Views.Build exposing (viewBuildStatusIconClasses, viewBuildTextClass)
import Views.Style as Style
import Util exposing ((=>))


-- EXTERNAL

import Html exposing (Html)
import Html.Styled.Attributes as Attributes exposing (css, class, classList)
import Html.Styled as Styled exposing (..)
import Css exposing (..)


-- CONFIG


type alias Config msg =
    { newUrlMsg : String -> msg }


type alias Context =
    { project : Project
    , builds : List Build
    , commit : Commit
    , tasks : List Task
    , selected : Maybe Task.Name
    , windowWidth : Int
    }


type alias NavTaskProperties =
    { isSelected : Bool
    , route : Route
    , iconClass : String
    , textClass : String
    , itemText : String
    }


type DisplayType
    = Visible
    | Hidden



-- VIEW


view : Config msg -> Context -> Html.Html msg
view config context =
    case displayType context.windowWidth of
        Visible ->
            div
                [ css
                    [ width (px 220)
                    , position fixed
                    , top (px 0)
                    , left (px 75)
                    , bottom (px 0)
                    , zIndex (int 1)
                    , backgroundColor (rgb 244 245 247)
                    , color (rgb 66 82 110)
                    ]
                ]
                [ details context.commit
                , taskNav config context
                ]
                |> toUnstyled

        Hidden ->
            toUnstyled (text "")


displayType : Int -> DisplayType
displayType windowWidth =
    if windowWidth >= 1024 then
        Visible
    else
        Hidden


details : Commit -> Styled.Html msg
details commit =
    div [ class "p-1" ]
        [ div [ class "card" ]
            [ div [ class "card-body" ]
                [ fromUnstyled (infoPanel commit)
                , hr [] []
                , fromUnstyled (branchList commit)
                , hr [] []
                , Styled.small [] [ text (truncateCommitMessage commit) ]
                ]
            ]
        ]


{-| List of task navigation
-}
taskNav : Config msg -> Context -> Styled.Html msg
taskNav config context =
    ul [ class "nav nav-pills flex-column project-navigation p-0" ] <|
        taskNavItems config context


taskNavItems : Config msg -> Context -> List (Styled.Html msg)
taskNavItems { newUrlMsg } context =
    context
        |> .tasks
        |> filterTasks
        |> sortTasks
        |> List.map (taskNavProperties context >> taskNavItem newUrlMsg)


{-| Single nav item for a task
-}
taskNavItem : (String -> msg) -> NavTaskProperties -> Styled.Html msg
taskNavItem newUrlMsg { isSelected, route, itemText, textClass, iconClass } =
    li [ class "nav-item" ]
        [ a
            [ class "nav-link text-secondary align-middle"
            , class textClass
            , css
                [ Style.textOverflowMixin
                , taskNavItemActiveCss isSelected
                , borderRadius (px 0)
                ]
            , Attributes.fromUnstyled (Route.href route)
            , Attributes.fromUnstyled (onClickPage newUrlMsg route)
            ]
            [ text itemText
            ]
        ]


taskNavItemActiveCss : Bool -> Style
taskNavItemActiveCss active =
    if active then
        Css.batch
            [ backgroundColor (hex "e2e3e5")
            , borderColor (hex "d6d8db")
            , color (hex "383d41")
            ]
    else
        Css.batch []



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
        |> Maybe.withDefault "fa-minus"


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
        |> List.reverse
        |> List.head


{-| Determine if a task is currently selected
-}
isSelected : Maybe Task.Name -> Task -> Bool
isSelected maybeTaskName task =
    case maybeTaskName of
        Just selected ->
            selected == task.name

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
