module Views.Page exposing (frame, ActivePage(..))

{-| The frame around a typical page - that is, the header and footer.
-}

import Html exposing (..)
import Html.Attributes exposing (..)
import Route exposing (Route)
import Data.User as User exposing (User, Username)
import Html.Lazy exposing (lazy2)
import Views.Spinner exposing (spinner)
import Views.Helpers exposing (onClickPage)


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
        [ viewContent content
        , viewFooter
        ]


viewContent : Html msg -> Html msg
viewContent content =
    div []
        [ content ]


viewFooter : Html msg
viewFooter =
    div [] []
