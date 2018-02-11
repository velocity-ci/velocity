module Page.Header exposing (ExternalMsg(..), viewHeader)

import Util exposing ((=>))
import Views.Page as Page exposing (ActivePage(..))
import Data.User as User exposing (User, Username)
import Html exposing (..)
import Html.Attributes exposing (..)
import Route exposing (Route)
import Views.Spinner exposing (spinner)
import Html.Lazy exposing (lazy2)
import Views.Helpers exposing (onClickPage)


type ExternalMsg
    = NewUrl String


viewHeader : Maybe User -> Bool -> ActivePage -> Html ExternalMsg
viewHeader user isLoading page =
    nav [ class "navbar navbar-expand-md navbar-dark fixed-top bg-dark" ]
        [ a
            [ class "navbar-brand"
            , onClickPage NewUrl Route.Home
            , Route.href Route.Home
            ]
            [ text "Velocity CI" ]
        , div [ class "collapse navbar-collapse", id "navbarCollapse" ]
            [ ul [ class "navbar-nav ml-auto" ] <|
                lazy2 Util.viewIf isLoading spinner
                    :: viewSignIn page user
            ]
        ]


viewSignIn : ActivePage -> Maybe User -> List (Html ExternalMsg)
viewSignIn page user =
    case user of
        Nothing ->
            []

        Just user ->
            [ navbarLink (page == Projects) Route.Projects [ text "Projects" ]
            , navbarLink (page == KnownHosts) Route.KnownHosts [ text "Known hosts" ]
            , navbarLink False Route.Logout [ text "Sign out" ]
            ]


navbarLink : Bool -> Route -> List (Html ExternalMsg) -> Html ExternalMsg
navbarLink isActive route linkContent =
    li [ classList [ ( "nav-item", True ), ( "active", isActive ) ] ]
        [ a
            [ class "nav-link"
            , Route.href route
            , onClickPage NewUrl route
            ]
            linkContent
        ]
