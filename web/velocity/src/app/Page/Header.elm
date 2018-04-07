module Page.Header exposing (Msg(..), view, Model, init, update, subscriptions)

import Util exposing ((=>))
import Views.Page as Page exposing (ActivePage(..))
import Data.User as User exposing (User, Username)
import Html exposing (..)
import Html.Attributes exposing (..)
import Route exposing (Route)
import Views.Spinner exposing (spinner)
import Views.Helpers exposing (onClickPage)
import Bootstrap.Navbar as Navbar
import Navigation
import Color


-- MODEL --


type alias Model =
    { navbarState : Navbar.State }


init : ( Model, Cmd Msg )
init =
    let
        ( navbarState, navBarCmd ) =
            Navbar.initialState NavbarMsg
    in
        ( { navbarState = navbarState }, navBarCmd )



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions model =
    Navbar.subscriptions model.navbarState NavbarMsg



-- VIEW --


view : Model -> Maybe User -> Bool -> ActivePage -> Html Msg
view model user isLoading page =
    Navbar.config NavbarMsg
        |> Navbar.withAnimation
        |> Navbar.collapseExtraLarge
        |> Navbar.lightCustom Color.white
        |> Navbar.fixTop
        |> Navbar.brand [ onClickPage NewUrl Route.Home, Route.href Route.Home ] [ text "Velocity CI" ]
        |> Navbar.items (navbarItems user page)
        |> Navbar.customItems (navbarCustomItems isLoading)
        |> Navbar.view model.navbarState


navbarItems : Maybe User -> ActivePage -> List (Navbar.Item Msg)
navbarItems user page =
    case user of
        Nothing ->
            []

        Just user ->
            [ navbarLink (page == Projects) Route.Projects [ text "Projects" ]
            , navbarLink (page == KnownHosts) Route.KnownHosts [ text "Known hosts" ]
            , navbarLink False Route.Logout [ text "Sign out" ]
            ]


navbarCustomItems : Bool -> List (Navbar.CustomItem Msg)
navbarCustomItems isLoading =
    [ navbarLoadingSpinner isLoading ]


navbarLink : Bool -> Route -> List (Html Msg) -> Navbar.Item Msg
navbarLink isActive route linkContent =
    if isActive then
        Navbar.itemLinkActive [ Route.href route, onClickPage NewUrl route ] linkContent
    else
        Navbar.itemLink [ Route.href route, onClickPage NewUrl route ] linkContent


navbarLoadingSpinner : Bool -> Navbar.CustomItem Msg
navbarLoadingSpinner isLoading =
    Navbar.textItem [] [ Util.viewIf isLoading spinner ]



--UPDATE --


type Msg
    = NavbarMsg Navbar.State
    | NewUrl String


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NavbarMsg state ->
            { model | navbarState = state } => Cmd.none

        NewUrl newUrl ->
            let
                ( navbarState, navBarCmd ) =
                    Navbar.initialState NavbarMsg
            in
                { model | navbarState = navbarState }
                    ! [ Navigation.newUrl newUrl, navBarCmd ]
