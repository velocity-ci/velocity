module Views.Page exposing (frame, sidebarFrame, ActivePage(..), SidebarType(..))

{-| The frame around a typical page - that is, the header and footer.
-}

import Html exposing (..)
import Html.Attributes exposing (class, classList)
import Data.User as User exposing (User)
import Route as Route
import Views.Helpers exposing (onClickPage)
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
    | Users


{-| Determines the amount of margin-left width to pad the main frame
-}
type SidebarType
    = NoSidebar
    | NormalSidebar
    | ExtraWideSidebar


{-| Take a page's Html and frame it with a header and footer.

The caller provides the current user, so we can display in either
"signed in" (rendering username) or "signed out" mode.

isLoading is for determining whether we should show a loading spinner
in the header. (This comes up during slow page transitions.)

-}
frame : Bool -> Maybe User -> ActivePage -> SidebarType -> Html msg -> Html msg
frame isLoading user page sidebarType content =
    div [ class "content-container px-4", class (sidebarClass sidebarType) ]
        [ viewContent content
        , viewFooter
        ]


sidebarClass : SidebarType -> String
sidebarClass sidebarType =
    case sidebarType of
        NoSidebar ->
            ""

        NormalSidebar ->
            "has-sidebar"

        ExtraWideSidebar ->
            "has-extra-sidebar"


viewContent : Html msg -> Html msg
viewContent content =
    div [] [ content ]


sidebarFrame : (String -> msg) -> Html msg -> Html msg
sidebarFrame newUrlMsg content =
    nav [ class "sidebar" ]
        [ sidebarLogo newUrlMsg
        , content
        ]


sidebarLogo : (String -> msg) -> Html msg
sidebarLogo newUrlMsg =
    div [ class "d-flex justify-content-center" ]
        [ a
            [ class "brand"
            , Route.href Route.Home
            , onClickPage newUrlMsg Route.Home
            ]
            [ h1 [] [ i [ class "fa fa-arrow-circle-o-right" ] [] ]
            ]
        ]


viewFooter : Html msg
viewFooter =
    div [] []
