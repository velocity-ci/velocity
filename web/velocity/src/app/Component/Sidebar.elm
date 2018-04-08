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
    { newUrlMsg : String -> msg }


type alias State a =
    { a
        | commitIconPopover : Popover.State
        , projectIconPopover : Popover.State
    }


view : Config msg -> Project -> ActiveSubPage -> Html msg
view config project subPage =
    nav [ class "sidebar bg-secondary" ]
        [ ul [ class "nav nav-pills flex-column" ]
            [ sidebarProjectLink config project
            , sidebarLink config
                (subPage == CommitsPage)
                (Route.Project project.slug (ProjectRoute.Commits Nothing Nothing))
                [ i [ attribute "aria-hidden" "true", class "fa fa-code-fork" ] [] ]
            , sidebarLink config
                (subPage == SettingsPage)
                (Route.Project project.slug ProjectRoute.Settings)
                [ i [ attribute "aria-hidden" "true", class "fa fa-cog" ] [] ]
            ]
        ]


sidebarProjectLink : Config msg -> Project -> Html msg
sidebarProjectLink config project =
    sidebarLink config
        False
        (Route.Project project.slug ProjectRoute.Overview)
        [ div
            [ class "badge badge-primary project-badge" ]
            [ i [ attribute "aria-hidden" "true", class "fa fa-code" ] []
            ]
        ]


sidebarLink : Config msg -> Bool -> Route -> List (Html msg) -> Html msg
sidebarLink { newUrlMsg } isActive route linkContent =
    li [ class "nav-item" ]
        [ a
            [ class "nav-link text-light text-center h4"
            , Route.href route
            , classList [ ( "active", isActive ) ]
            , onClickPage newUrlMsg route
            ]
            linkContent
        ]


popover :
    (Popover.Config msg -> Popover.Config msg1)
    -> Popover.State
    -> Html msg
    -> Html msg1
popover posFn popState btn =
    Popover.config btn
        |> posFn
        |> Popover.content [] [ text "Tooltip" ]
        |> Popover.view popState
