module Views.Page exposing (frame, sidebarFrame, ActivePage(..), SidebarType(..))

{-| The frame around a typical page - that is, the header and footer.
-}

import Html exposing (Html)
import Html.Styled.Attributes as Attributes exposing (css, class, classList)
import Html.Styled as Styled exposing (..)
import Css exposing (..)
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
frame : Bool -> Maybe User -> ActivePage -> SidebarType -> Html.Html msg -> Html.Html msg
frame isLoading user page sidebarType content =
    div
        [ class "px-4"
        , css
            [ marginLeft (px <| sidebarWidthPx sidebarType) ]
        ]
        [ viewContent content
        , viewFooter
        ]
        |> toUnstyled


sidebarWidthPx : SidebarType -> Float
sidebarWidthPx sidebarType =
    case sidebarType of
        NoSidebar ->
            0

        NormalSidebar ->
            75

        ExtraWideSidebar ->
            295


viewContent : Html.Html msg -> Styled.Html msg
viewContent content =
    div [] [ fromUnstyled content ]


sidebarFrame : (String -> msg) -> Html.Html msg -> Html.Html msg
sidebarFrame newUrlMsg content =
    nav
        [ css
            [ width (px 75)
            , position fixed
            , top (px 0)
            , bottom (px 0)
            , left (px 0)
            , zIndex (int 2)
            , paddingTop (Css.rem 1)
            , backgroundColor (rgb 7 71 166)
            , color (hex "ffffff")
            ]
        ]
        [ sidebarLogo newUrlMsg
        , fromUnstyled content
        ]
        |> toUnstyled


sidebarLogo : (String -> msg) -> Styled.Html msg
sidebarLogo newUrlMsg =
    div [ class "d-flex justify-content-center" ]
        [ a
            [ css
                [ color (hex "ffffff")
                , hover
                    [ color (hex "ffffff") ]
                ]
            , Attributes.fromUnstyled (Route.href Route.Home)
            , Attributes.fromUnstyled (onClickPage newUrlMsg Route.Home)
            ]
            [ h1 [] [ i [ class "fa fa-arrow-circle-o-right" ] [] ]
            ]
        ]


viewFooter : Styled.Html msg
viewFooter =
    div [] []
