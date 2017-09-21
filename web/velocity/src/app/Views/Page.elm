module Views.Page exposing (frame, ActivePage(..))

{-| The frame around a typical page - that is, the header and footer.
-}

import Html exposing (..)
import Html.Attributes exposing (..)
import Route exposing (Route)
import Data.User as User exposing (User, Username)
import Html.Lazy exposing (lazy2)
import Views.Spinner exposing (spinner)
import Util exposing ((=>))


{-| Determines which navbar link (if any) will be rendered as active.

Note that we don't enumerate every page here, because the navbar doesn't
have links for every page. Anything that's not part of the navbar falls
under Other.

-}
type ActivePage
    = Other
    | Home
    | Login
    | Projects
    | Project
    | KnownHosts


{-| Take a page's Html and frame it with a header and footer.

The caller provides the current user, so we can display in either
"signed in" (rendering username) or "signed out" mode.

isLoading is for determining whether we should show a loading spinner
in the header. (This comes up during slow page transitions.)

-}
frame : Bool -> Maybe User -> ActivePage -> Html msg -> Html msg
frame isLoading user page content =
    div []
        [ viewHeader page user isLoading
        , viewContent content
        , viewFooter
        ]


viewHeader : ActivePage -> Maybe User -> Bool -> Html msg
viewHeader page user isLoading =
    nav [ class "navbar navbar-expand-md navbar-dark fixed-top bg-dark" ]
        [ a [ class "navbar-brand", Route.href Route.Home ]
            [ text "VeloCIty" ]
        , div [ class "collapse navbar-collapse", id "navbarCollapse" ]
            [ ul [ class "navbar-nav ml-auto" ] <|
                lazy2 Util.viewIf isLoading spinner
                    :: viewSignIn page user
            ]
        ]


viewContent : Html msg -> Html msg
viewContent content =
    div []
        [ content ]


viewSignIn : ActivePage -> Maybe User -> List (Html msg)
viewSignIn page user =
    case user of
        Nothing ->
            [ navbarLink (page == Login) Route.Login [ text "Sign in" ]
            ]

        Just user ->
            [ navbarLink (page == Projects) Route.Projects [ text "Projects" ]
            , navbarLink (page == KnownHosts) Route.KnownHosts [ text "Known hosts" ]
            , navbarLink False Route.Logout [ text "Sign out" ]
            ]


viewFooter : Html msg
viewFooter =
    div [] []


navbarLink : Bool -> Route -> List (Html msg) -> Html msg
navbarLink isActive route linkContent =
    li [ classList [ ( "nav-item", True ), ( "active", isActive ) ] ]
        [ a [ class "nav-link", Route.href route ] linkContent ]
