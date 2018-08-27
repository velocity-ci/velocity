module Views.Page exposing (frame, sidebarFrame, ActivePage(..))

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
import Component.Sidebar as Sidebar


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


{-| Take a page's Html and frame it with a header and footer.

The caller provides the current user, so we can display in either
"signed in" (rendering username) or "signed out" mode.

isLoading is for determining whether we should show a loading spinner
in the header. (This comes up during slow page transitions.)

-}
frame : Bool -> Maybe User -> Sidebar.DisplayType -> ActivePage -> Html.Html msg -> Html.Html msg
frame isLoading user sidebarType page content =
    div
        [ css
            [ marginLeft (px <| Sidebar.sidebarWidth sidebarType) ]
        ]
        [ viewContent content
        , viewFooter
        ]
        |> toUnstyled


viewContent : Html.Html msg -> Styled.Html msg
viewContent content =
    div [] [ fromUnstyled content ]


sidebarFrame : Sidebar.DisplayType -> Sidebar.Config msg -> Html.Html msg -> Html.Html msg -> Html.Html msg
sidebarFrame displayType sidebarConfig sidebarContent subSidebarContent =
    nav
        [ css
            [ width (px <| Sidebar.sidebarWidth displayType)
            , position fixed
            , top (px 0)
            , bottom (px 0)
            , left (px <| Sidebar.sidebarLeft displayType)
            , zIndex (int 2)
            , paddingTop (Css.rem 1)
            , backgroundColor (rgb 7 71 166)
            , color (hex "ffffff")
            ]
        ]
        [ sidebarLogo sidebarConfig.newUrlMsg
        , fromUnstyled sidebarContent
        ]
        |> toUnstyled



--
--subSidebarFrame : Sidebar.Config msg -> Sidebar.DisplayType -> Html.Html msg -> Html.Html msg
--subSidebarFrame config displayType content =
--    Sidebar.view config displayType content


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
