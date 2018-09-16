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
        []
        [ Util.viewIfStyled (Sidebar.isCollapsable sidebarType) (viewNavbar sidebarConfig)
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
                    [ (Sidebar.sidebarAnimationAttrs sidebarDisplayType)
                    , [ css
                            [ position relative
                            , width (pct 100)
                            ]
                      ]
                    ]
                )
                [ fromUnstyled content ]
            ]


viewNavbar : Sidebar.Config msg -> Styled.Html msg
viewNavbar { toggleSidebarMsg } =
    nav
        [ class "navbar navbar-light bg-light border-bottom"
        , css
            [ borderBottomColor (hex "d4dadf")
            ]
        ]
        [ viewNavbarToggle toggleSidebarMsg ]


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
                , [ class "d-flex border-right"
                  , css
                        [ width (px <| Sidebar.sidebarWidth displayType)
                        , position fixed
                        , borderRightColor (hex "d4dadf")
                        , top
                            (px
                                (if Sidebar.isCollapsable displayType then
                                    56
                                 else
                                    0
                                )
                            )
                        , bottom (px 0)
                        , zIndex (int 2)
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
