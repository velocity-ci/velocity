module Component.Sidebar exposing (Config, ActiveSubPage(..), view)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Bootstrap.Popover as Popover


-- INTERNAL --

import Data.Project as Project exposing (Project)
import Page.Project.Route as ProjectRoute
import Route exposing (Route)
import Views.Helpers exposing (onClickPage)


-- VIEW --


type ActiveSubPage
    = OtherPage
    | OverviewPage
    | CommitsPage
    | SettingsPage


type alias Config msg =
    { newUrlMsg : String -> msg
    , commitPopMsg : Popover.State -> msg
    , settingsPopMsg : Popover.State -> msg
    }


type alias State a =
    { a
        | commitIconPopover : Popover.State
        , settingsIconPopover : Popover.State
    }


view : State a -> Config msg -> Project -> ActiveSubPage -> Html msg
view state config project subPage =
    nav [ class "sidebar bg-secondary" ]
        [ ul [ class "nav nav-pills flex-column" ]
            [ sidebarProjectLink state config project
            , sidebarLink state
                config
                CommitsPage
                (subPage == CommitsPage)
                (Route.Project project.slug (ProjectRoute.Commits Nothing Nothing))
                "Commits"
                [ i [ attribute "aria-hidden" "true", class "fa fa-code-fork" ] [] ]
            , sidebarLink state
                config
                SettingsPage
                (subPage == SettingsPage)
                (Route.Project project.slug ProjectRoute.Settings)
                "Settings"
                [ i [ attribute "aria-hidden" "true", class "fa fa-cog" ] [] ]
            ]
        ]


sidebarProjectLink : State a -> Config msg -> Project -> Html msg
sidebarProjectLink state config project =
    sidebarLink state
        config
        OverviewPage
        False
        (Route.Project project.slug ProjectRoute.Overview)
        "Overview"
        [ div
            [ class "badge badge-primary project-badge" ]
            [ i [ attribute "aria-hidden" "true", class "fa fa-code" ] []
            ]
        ]


sidebarLink : State a -> Config msg -> ActiveSubPage -> Bool -> Route -> String -> List (Html msg) -> Html msg
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
    li [ class "nav-item" ]
        [ a
            [ class "nav-link text-light text-center h4"
            , Route.href route
            , classList [ ( "active", isActive ) ]
            , onClickPage config.newUrlMsg route
            ]
            content
        ]


tooltipLink : Config msg -> Bool -> Route -> List (Html msg) -> ( Popover.State -> msg, Popover.State ) -> Html msg
tooltipLink config isActive route content ( popMsg, popState ) =
    li ([ class "nav-item" ] ++ Popover.onHover popState popMsg)
        [ a
            ([ class "nav-link text-light text-center h4"
             , Route.href route
             , classList [ ( "active", isActive ) ]
             , onClickPage config.newUrlMsg route
             ]
            )
            content
        ]


tooltipConfig : Config msg -> State a -> ActiveSubPage -> Maybe ( Popover.State -> msg, Popover.State )
tooltipConfig { commitPopMsg, settingsPopMsg } { commitIconPopover, settingsIconPopover } activeSubPage =
    case activeSubPage of
        CommitsPage ->
            Just ( commitPopMsg, commitIconPopover )

        SettingsPage ->
            Just ( settingsPopMsg, settingsIconPopover )

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
