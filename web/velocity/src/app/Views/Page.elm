module Views.Page exposing (ActivePage(..), frame, sidebar, sidebarFrame)

{-| The frame around a typical page - that is, the header and footer.
-}

import Component.Sidebar as Sidebar
import Css exposing (..)
import Data.User as User exposing (User)
import Html exposing (Html)
import Html.Styled as Styled exposing (..)
import Html.Styled.Attributes as Attributes exposing (class, classList, css, id)
import Html.Styled.Events exposing (onClick)
import Route as Route
import Util exposing ((=>))
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
    | Users


type alias SidebarConfigs msg =
    ( Sidebar.Config msg, Sidebar.Config msg )


type alias SidebarData =
    ( Sidebar.DisplayType, Sidebar.DisplayType )


{-| Take a page's Html and frame it with a header and footer.

The caller provides the current user, so we can display in either
"signed in" (rendering username) or "signed out" mode.

isLoading is for determining whether we should show a loading spinner
in the header. (This comes up during slow page transitions.)

-}
frame : Bool -> Maybe User -> SidebarConfigs msg -> SidebarData -> ActivePage -> Html.Html msg -> Html.Html msg
frame isLoading user sidebarConfigs ( sidebarType, subSidebarType ) page content =
    div
        []
        [ Util.viewIfStyled (Sidebar.isCollapsable sidebarType || Sidebar.isCollapsable subSidebarType) (viewNavbar sidebarConfigs)
        , viewContent sidebarType content
        , viewFooter
        ]
        |> toUnstyled


viewContent : Sidebar.DisplayType -> Html.Html msg -> Styled.Html msg
viewContent sidebarDisplayType content =
    let
        sidebarWidth =
            if Sidebar.isCollapsable sidebarDisplayType then
                calc (pct 100) plus (px <| Sidebar.sidebarWidth sidebarDisplayType)
            else
                calc (pct 100) plus (px 0)
    in
    div
        [ css
            [ paddingLeft (px (Sidebar.sidebarWidth sidebarDisplayType))
            , width sidebarWidth
            ]
        ]
        [ div
            (List.concat
                [ Sidebar.sidebarAnimationAttrs sidebarDisplayType
                , [ css
                        [ position relative
                        , width (pct 100)
                        ]
                  ]
                ]
            )
            [ fromUnstyled content ]
        ]


viewNavbar : SidebarConfigs msg -> Styled.Html msg
viewNavbar ( sidebarConfig, subSidebarConfig ) =
    nav
        [ class "navbar navbar-light bg-light border-bottom d-flex"
        , css
            [ borderBottomColor (hex "d4dadf")
            ]
        ]
        [ viewNavbarToggle sidebarConfig.toggleSidebarMsg
        , viewNavbarToggle subSidebarConfig.toggleSidebarMsg
        ]


viewNavbarClose : msg -> Styled.Html msg
viewNavbarClose showCollapsableSidebarMsg =
    button
        [ class "navbar-toggler"
        , onClick showCollapsableSidebarMsg
        ]
        [ i [ class "fa fa-arrow-left" ] []
        ]


viewNavbarToggle : msg -> Styled.Html msg
viewNavbarToggle toggleNavbarMsg =
    button
        [ class "navbar-toggler"
        , onClick toggleNavbarMsg
        ]
        [ span [ class "fa fa-bars" ] []
        ]


sidebar : List (Html.Html msg) -> Html.Html msg
sidebar items =
    div
        []
        (List.map fromUnstyled items)
        |> toUnstyled


sidebarFrame : Sidebar.DisplayType -> Sidebar.Config msg -> Html.Html msg -> Html.Html msg -> Html.Html msg
sidebarFrame displayType sidebarConfig sidebarContent subSidebarContent =
    div
        []
        [ div
            [ id "sidebar-overlay"
            , css (Sidebar.collapsableOverlay displayType)
            , onClick sidebarConfig.hideCollapsableSidebarMsg
            ]
            []
        , nav
            (List.concat
                [ Sidebar.sidebarAnimationAttrs displayType
                , [ class "border-right"
                  , css
                        [ width (px <| Sidebar.sidebarWidth displayType)
                        , position fixed
                        , borderRightColor (hex "d4dadf")
                        , top (px 0)
                        , bottom (px 0)
                        , zIndex (int 2)
                        , backgroundColor (rgb 7 71 166)
                        , color (hex "ffffff")
                        ]
                  ]
                ]
            )
            [ div
                [ class "navbar navbar-light bg-light border-bottom border-right-0"
                , css
                    [ borderBottomColor (hex "dee2e6") ]
                ]
                [ viewNavbarClose sidebarConfig.hideCollapsableSidebarMsg ]
            , div [ class "d-flex align-items-stretch", css [ height (pct 100) ] ]
                [ fromUnstyled sidebarContent
                , fromUnstyled subSidebarContent
                ]
            ]
        ]
        |> toUnstyled


viewFooter : Styled.Html msg
viewFooter =
    div [] []
