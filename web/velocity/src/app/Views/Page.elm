module Views.Page exposing (frame, sidebar, sidebarFrame, ActivePage(..))

{-| The frame around a typical page - that is, the header and footer.
-}

import Html exposing (Html)
import Html.Styled.Attributes as Attributes exposing (id, css, class, classList)
import Html.Styled as Styled exposing (..)
import Html.Styled.Events exposing (onClick)
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
frame : Bool -> Maybe User -> Sidebar.Config msg -> Sidebar.DisplayType -> ActivePage -> Html.Html msg -> Html.Html msg
frame isLoading user sidebarConfig sidebarType page content =
    div
        [ css
            [ marginLeft (px <| Sidebar.containerMarginLeft sidebarType) ]
        ]
        [ Util.viewIfStyled (Sidebar.isCollapsable sidebarType) (viewNavbar sidebarConfig)
        , viewContent content
        , viewFooter
        ]
        |> toUnstyled


viewContent : Html.Html msg -> Styled.Html msg
viewContent content =
    div [] [ fromUnstyled content ]


viewNavbar : Sidebar.Config msg -> Styled.Html msg
viewNavbar { showCollapsableSidebarMsg } =
    nav
        [ class "navbar navbar-light bg-light" ]
        [ viewNavbarToggle showCollapsableSidebarMsg ]


viewNavbarToggle : msg -> Styled.Html msg
viewNavbarToggle showCollapsableSidebarMsg =
    button
        [ class "navbar-toggler"
        , onClick showCollapsableSidebarMsg
        ]
        [ span [ class "navbar-toggler-icon" ] []
        ]


sidebar : List (Html.Html msg) -> Html.Html msg
sidebar items =
    div
        [ css [ width (px 75) ] ]
        (List.map fromUnstyled items)
        |> toUnstyled


sidebarFrame : Sidebar.DisplayType -> Sidebar.Config msg -> Html.Html msg -> Html.Html msg -> Html.Html msg
sidebarFrame displayType sidebarConfig sidebarContent subSidebarContent =
    div []
        [ div
            [ id "sidebar-overlay"
            , css (Sidebar.collapsableOverlay displayType)
            , onClick sidebarConfig.hideCollapsableSidebarMsg
            ]
            []
        , nav
            (List.concat
                [ Sidebar.sidebarAnimationAttrs displayType
                , [ class "d-flex justify-content-center"
                  , css
                        [ width (px <| Sidebar.sidebarWidth displayType)
                        , position fixed
                        , top (px 0)
                        , bottom (px 0)
                        , zIndex (int 2)
                        , paddingTop (Css.rem 1)
                        , backgroundColor (rgb 7 71 166)
                        , color (hex "ffffff")
                        ]
                  ]
                ]
            )
            [ fromUnstyled sidebarContent
            , fromUnstyled subSidebarContent
            ]
        ]
        |> toUnstyled


viewFooter : Styled.Html msg
viewFooter =
    div [] []
