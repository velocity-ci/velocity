module Component.CommitSidebar
    exposing
        ( view
        , animate
        , show
        , hide
        , subscriptions
        , Context
        , Config
        , initDisplayType
        , fixedVisible
        , collapsableVisible
        , collapsableHidden
        , DisplayType
        )

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
import Html.Styled.Events exposing (onClick)
import Css exposing (..)
import Animation


-- CONFIG


type alias Config msg =
    { newUrlMsg : String -> msg
    , animateMsg : Animation.Msg -> msg
    , hideCollapsableSidebarMsg : msg
    }


type alias Context =
    { project : Project
    , builds : List Build
    , commit : Commit
    , tasks : List Task
    , selected : Maybe Task.Name
    , displayType : DisplayType
    }


type alias NavTaskProperties =
    { isSelected : Bool
    , route : Route
    , iconClass : String
    , textClass : String
    , itemText : String
    }


type DisplayType
    = Fixed
    | Collapsable CollapsableVisibility


type CollapsableVisibility
    = Visible Animation.State
    | Hidden Animation.State



-- SUBSCRIPTIONS --


subscriptions : Config msg -> Context -> Sub msg
subscriptions { animateMsg } { displayType } =
    case displayType of
        Collapsable (Visible animationState) ->
            Animation.subscription animateMsg [ animationState ]

        Collapsable (Hidden animationState) ->
            Animation.subscription animateMsg [ animationState ]

        _ ->
            Sub.none



-- UPDATE --


show : DisplayType -> DisplayType
show displayType =
    case displayType of
        Collapsable (Hidden animationState) ->
            animationState
                |> Animation.interrupt [ Animation.to animationFinishAttrs ]
                |> Visible
                |> Collapsable

        _ ->
            displayType


hide : DisplayType -> DisplayType
hide displayType =
    case displayType of
        Collapsable (Visible animationState) ->
            animationState
                |> Animation.interrupt [ Animation.to animationStartAttrs ]
                |> Hidden
                |> Collapsable

        _ ->
            displayType


animate : DisplayType -> Animation.Msg -> DisplayType
animate displayType msg =
    case displayType of
        Collapsable (Visible animationState) ->
            animationState
                |> Animation.update msg
                |> Visible
                |> Collapsable

        Collapsable (Hidden animationState) ->
            animationState
                |> Animation.update msg
                |> Hidden
                |> Collapsable

        _ ->
            displayType



-- VIEW --


view : Config msg -> Context -> Html.Html msg
view config context =
    context
        |> sidebarContainer config
        |> toUnstyled


fixedVisible : DisplayType
fixedVisible =
    Fixed


collapsableVisible : DisplayType
collapsableVisible =
    Collapsable (Visible <| Animation.style animationFinishAttrs)


collapsableHidden : DisplayType
collapsableHidden =
    Collapsable (Hidden <| Animation.style animationStartAttrs)


animationStartAttrs : List Animation.Property
animationStartAttrs =
    [ Animation.left (Animation.px -145.0) ]


animationFinishAttrs : List Animation.Property
animationFinishAttrs =
    [ Animation.left (Animation.px 75.0) ]


sidebarContainer : Config msg -> Context -> Styled.Html msg
sidebarContainer config context =
    div []
        [ div
            [ css (collapsableOverlay context.displayType)
            , onClick config.hideCollapsableSidebarMsg
            ]
            []
        , sidebar config context
        ]


sidebar : Config msg -> Context -> Styled.Html msg
sidebar config context =
    div
        (List.concat
            [ sidebarAnimationAttrs context.displayType
            , [ css
                    [ sidebarBaseStyle
                    , sidebarStyle context.displayType
                    ]
              ]
            ]
        )
        [ details context.commit
        , taskNav config context
        ]


collapsableOverlay : DisplayType -> List Style
collapsableOverlay displayType =
    case displayType of
        Collapsable (Visible _) ->
            [ position fixed
            , top (px 0)
            , right (px 0)
            , left (px 75)
            , bottom (px 0)
            , zIndex (int 1)
            , backgroundColor (hex "000000")
            , opacity (num 0.5)
            ]

        _ ->
            [ display none ]


sidebarAnimationAttrs : DisplayType -> List (Attribute msg)
sidebarAnimationAttrs displayType =
    case displayType of
        Collapsable (Visible animationState) ->
            animationToStyledAttrs animationState

        Collapsable (Hidden animationState) ->
            animationToStyledAttrs animationState

        Fixed ->
            []


animationToStyledAttrs : Animation.State -> List (Attribute msg)
animationToStyledAttrs animationState =
    animationState
        |> Animation.render
        |> List.map Attributes.fromUnstyled


sidebarStyle : DisplayType -> Style
sidebarStyle displayType =
    case displayType of
        Fixed ->
            width (px 220)

        Collapsable _ ->
            width (px 220)


sidebarBaseStyle : Style
sidebarBaseStyle =
    Css.batch
        [ top (px 0)
        , left (px 75)
        , bottom (px 0)
        , zIndex (int 1)
        , backgroundColor (rgb 244 245 247)
        , color (rgb 66 82 110)
        , position fixed
        ]


initDisplayType : Int -> DisplayType
initDisplayType windowWidth =
    if windowWidth >= 992 then
        fixedVisible
    else
        collapsableHidden


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
