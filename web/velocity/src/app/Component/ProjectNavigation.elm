module Component.ProjectNavigation exposing (ActiveSubPage(..), Config, State, view)

-- EXTERNAL --
-- INTERNAL --

import Bootstrap.Popover as Popover
import Css exposing (..)
import Data.Project as Project exposing (Project)
import Html exposing (Html)
import Html.Styled as Styled exposing (..)
import Html.Styled.Attributes as StyledAttributes exposing (..)
import Page.Project.Route as ProjectRoute
import Route exposing (Route)
import Views.Helpers exposing (styledOnClickPage)
import Views.Project exposing (badge)


-- MODEL --


type ActiveSubPage
    = OtherPage
    | OverviewPage
    | CommitsPage
    | SettingsPage
    | BuildsPage


type alias Config msg =
    { newUrlMsg : String -> msg
    , commitPopMsg : Popover.State -> msg
    , buildsPopMsg : Popover.State -> msg
    , settingsPopMsg : Popover.State -> msg
    , projectBadgePopMsg : Popover.State -> msg
    }


type alias State =
    { commitIconPopover : Popover.State
    , buildsIconPopover : Popover.State
    , settingsIconPopover : Popover.State
    , projectBadgePopover : Popover.State
    }



-- VIEW --


view : State -> Config msg -> Project -> ActiveSubPage -> Html.Html msg
view state config project subPage =
    subPage
        |> sidebarProjectNavigation state config project
        |> toUnstyled


sidebarProjectNavigation : State -> Config msg -> Project -> ActiveSubPage -> Styled.Html msg
sidebarProjectNavigation state config project subPage =
    ul [ class "nav nav-pills flex-column project-navigation" ]
        [ sidebarProjectLink state config project
        , sidebarLink state
            config
            CommitsPage
            (subPage == CommitsPage)
            (Route.Project project.slug <| ProjectRoute.Commits Nothing Nothing)
            "Project commits"
            [ i [ attribute "aria-hidden" "true", class "fa fa-code-fork" ] [] ]
        , sidebarLink state
            config
            BuildsPage
            (subPage == BuildsPage)
            (Route.Project project.slug <| ProjectRoute.Builds Nothing)
            "Project builds"
            [ i [ attribute "aria-hidden" "true", class "fa fa-list-alt" ] [] ]
        , sidebarLink state
            config
            SettingsPage
            (subPage == SettingsPage)
            (Route.Project project.slug ProjectRoute.Settings)
            "Project settings"
            [ i [ attribute "aria-hidden" "true", class "fa fa-wrench" ] [] ]
        ]


sidebarProjectLink : State -> Config msg -> Project -> Styled.Html msg
sidebarProjectLink state config project =
    sidebarLink state
        config
        OverviewPage
        False
        (Route.Project project.slug ProjectRoute.Overview)
        project.name
        [ Styled.fromUnstyled (badge project) ]


sidebarLink : State -> Config msg -> ActiveSubPage -> Bool -> Route -> String -> List (Styled.Html msg) -> Styled.Html msg
sidebarLink state config activeSubPage isActive route tooltip linkContent =
    tooltipConfig config state activeSubPage
        |> Maybe.map
            (\( popMsg, popState ) ->
                tooltipLink config isActive route linkContent ( popMsg, popState )
                    |> popover Popover.right popState tooltip
            )
        |> Maybe.withDefault (nonTooltipLink config isActive route linkContent)


nonTooltipLink : Config msg -> Bool -> Route -> List (Styled.Html msg) -> Styled.Html msg
nonTooltipLink config isActive route content =
    li []
        [ a
            [ Route.styledHref route
            , classList [ ( "active", isActive ) ]
            , styledOnClickPage config.newUrlMsg route
            ]
            content
        ]


tooltipLink : Config msg -> Bool -> Route -> List (Styled.Html msg) -> ( Popover.State -> msg, Popover.State ) -> Styled.Html msg
tooltipLink config isActive route content ( popMsg, popState ) =
    li
        ([ class "nav-item"
         , css
            [ borderRadius (px 0)
            , hover [ backgroundColor (hex "ffffff") ]
            ]
         ]
            ++ (Popover.onHover popState popMsg |> List.map StyledAttributes.fromUnstyled)
        )
        [ a
            [ class "nav-link text-center h4"
            , Route.styledHref route
            , classList [ ( "active", isActive ) ]
            , styledOnClickPage config.newUrlMsg route
            ]
            content
        ]


tooltipConfig : Config msg -> State -> ActiveSubPage -> Maybe ( Popover.State -> msg, Popover.State )
tooltipConfig config state activeSubPage =
    case activeSubPage of
        CommitsPage ->
            Just ( config.commitPopMsg, state.commitIconPopover )

        BuildsPage ->
            Just ( config.buildsPopMsg, state.buildsIconPopover )

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
    -> Styled.Html msg
    -> Styled.Html msg1
popover posFn popState tooltipText btn =
    Popover.config (toUnstyled btn)
        |> posFn
        |> Popover.content [] [ toUnstyled <| text tooltipText ]
        |> Popover.view popState
        |> Styled.fromUnstyled
