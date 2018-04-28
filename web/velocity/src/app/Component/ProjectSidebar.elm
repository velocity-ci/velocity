module Component.ProjectSidebar exposing (State, Config, ActiveSubPage(..), view)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Bootstrap.Popover as Popover
import Bootstrap.Button as Button


-- INTERNAL --

import Data.Project as Project exposing (Project)
import Page.Project.Route as ProjectRoute
import Route exposing (Route)
import Views.Helpers exposing (onClickPage)
import Views.Project exposing (badge)


-- MODEL --


type ActiveSubPage
    = OtherPage
    | OverviewPage
    | CommitsPage
    | SettingsPage


type alias Config msg =
    { newUrlMsg : String -> msg
    , commitPopMsg : Popover.State -> msg
    , settingsPopMsg : Popover.State -> msg
    , projectBadgePopMsg : Popover.State -> msg
    }


type alias State =
    { commitIconPopover : Popover.State
    , settingsIconPopover : Popover.State
    , projectBadgePopover : Popover.State
    }



-- VIEW --


view : State -> Config msg -> Project -> ActiveSubPage -> Html msg
view state config project subPage =
    sidebarProjectNavigation state config project subPage


sidebarProjectNavigation : State -> Config msg -> Project -> ActiveSubPage -> Html msg
sidebarProjectNavigation state config project subPage =
    ul [ class "nav nav-pills flex-column project-navigation" ]
        [ sidebarProjectLink state config project
        , sidebarLink state
            config
            CommitsPage
            (subPage == CommitsPage)
            (Route.Project project.slug (ProjectRoute.Commits Nothing Nothing))
            "Project commits"
            [ i [ attribute "aria-hidden" "true", class "fa fa-code-fork" ] [] ]
        , sidebarLink state
            config
            SettingsPage
            (subPage == SettingsPage)
            (Route.Project project.slug ProjectRoute.Settings)
            "Project settings"
            [ i [ attribute "aria-hidden" "true", class "fa fa-wrench" ] [] ]
        ]


sidebarProjectLink : State -> Config msg -> Project -> Html msg
sidebarProjectLink state config project =
    sidebarLink state
        config
        OverviewPage
        False
        (Route.Project project.slug ProjectRoute.Overview)
        project.name
        [ badge ]


sidebarLink : State -> Config msg -> ActiveSubPage -> Bool -> Route -> String -> List (Html msg) -> Html msg
sidebarLink state config activeSubPage isActive route tooltip linkContent =
    tooltipConfig config state activeSubPage
        |> Maybe.map
            (\( popMsg, popState ) ->
                tooltipLink config isActive route linkContent ( popMsg, popState )
                    |> popover Popover.right popState tooltip
            )
        |> Maybe.withDefault (nonTooltipLink config isActive route linkContent)


nonTooltipLink : Config msg -> Bool -> Route -> List (Html msg) -> Html msg
nonTooltipLink config isActive route content =
    li []
        [ a
            [ Route.href route
            , classList [ ( "active", isActive ) ]
            , onClickPage config.newUrlMsg route
            ]
            content
        ]


tooltipLink : Config msg -> Bool -> Route -> List (Html msg) -> ( Popover.State -> msg, Popover.State ) -> Html msg
tooltipLink config isActive route content ( popMsg, popState ) =
    li ([ class "nav-item" ] ++ Popover.onHover popState popMsg)
        [ a
            ([ class "nav-link text-center h4"
             , Route.href route
             , classList [ ( "active", isActive ) ]
             , onClickPage config.newUrlMsg route
             ]
            )
            content
        ]


tooltipConfig : Config msg -> State -> ActiveSubPage -> Maybe ( Popover.State -> msg, Popover.State )
tooltipConfig config state activeSubPage =
    case activeSubPage of
        CommitsPage ->
            Just ( config.commitPopMsg, state.commitIconPopover )

        SettingsPage ->
            Just ( config.settingsPopMsg, state.settingsIconPopover )

        OverviewPage ->
            Just ( config.projectBadgePopMsg, state.projectBadgePopover )

        _ ->
            Nothing


popover :
    (Popover.Config msg -> Popover.Config msg1)
    -> Popover.State
    -> String
    -> Html msg
    -> Html msg1
popover posFn popState tooltipText btn =
    Popover.config btn
        |> posFn
        |> Popover.content [] [ text tooltipText ]
        |> Popover.view popState
