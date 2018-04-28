module Component.UserSidebar exposing (Config, State, init, view, subscriptions)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Button as Button


-- INTERNAL --

import Data.Project as Project exposing (Project)
import Page.Project.Route as ProjectRoute
import Route exposing (Route)
import Views.Helpers exposing (onClickPage)
import Views.Project exposing (badge)


-- MODEL --


type alias Config msg =
    { newUrlMsg : String -> msg
    , userDropdownMsg : Dropdown.State -> msg
    }


type alias State =
    { userDropdown : Dropdown.State }


init : State
init =
    { userDropdown = Dropdown.initialState }



-- SUBSCRIPTIONS --


subscriptions : Config msg -> State -> Sub msg
subscriptions { userDropdownMsg } { userDropdown } =
    Dropdown.subscriptions userDropdown userDropdownMsg



-- VIEW --


view : State -> Config msg -> Html msg
view state config =
    div [] [ sidebarUserDropdown state config ]


sidebarUserDropdown : State -> Config msg -> Html msg
sidebarUserDropdown { userDropdown } config =
    Dropdown.dropdown
        userDropdown
        { options =
            [ Dropdown.dropUp
            , Dropdown.attrs [ class "menu-toggle-dropdown d-flex justify-content-center" ]
            ]
        , toggleMsg = config.userDropdownMsg
        , toggleButton =
            Dropdown.toggle
                [ Button.light
                , Button.large
                ]
                []
        , items =
            [ Dropdown.header [ text "Management" ]
            , Dropdown.buttonItem [ onClickPage config.newUrlMsg Route.KnownHosts ] [ text "Known hosts" ]
            , Dropdown.buttonItem [ onClickPage config.newUrlMsg Route.Projects ] [ text "Projects" ]
            , Dropdown.header [ text "User" ]
            , Dropdown.buttonItem [ onClickPage config.newUrlMsg Route.Logout ] [ text "Log out" ]
            ]
        }
