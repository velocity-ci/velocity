module Views.Page exposing (frame, sidebarFrame, ActivePage(..))

{-| The frame around a typical page - that is, the header and footer.
-}

import Html exposing (..)
import Html.Attributes exposing (class)
import Data.User as User exposing (User)


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
    div [ class "content-container px-4" ]
        [ viewContent content
        , viewFooter
        ]


viewContent : Html msg -> Html msg
viewContent content =
    div []
        [ content ]


sidebarFrame : Html msg -> Html msg
sidebarFrame content =
    nav [ class "sidebar" ]
        [ content ]


viewFooter : Html msg
viewFooter =
    div [] []
