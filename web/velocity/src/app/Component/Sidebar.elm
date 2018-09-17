module Component.Sidebar
    exposing
        ( Config
        , Direction(..)
        , DisplayType
        , Size
        , animate
        , collapsableOverlay
        , collapsableVisible
        , extraWideSize
        , fixedVisibleExtraWide
        , hide
        , initDisplayType
        , isCollapsable
        , normalSize
        , show
        , sidebarAnimationAttrs
        , sidebarWidth
        , subscriptions
        , toggle
        , view
        )

-- INTERNAL
-- EXTERNAL

import Animation
import Css exposing (..)
import Data.Build as Build exposing (Build)
import Data.Commit as Commit exposing (Commit)
import Data.Device as Device
import Data.Project as Project exposing (Project)
import Data.Task as Task exposing (Task)
import Html exposing (Html)
import Html.Styled as Styled exposing (..)
import Html.Styled.Attributes as Attributes exposing (class, classList, css)
import Html.Styled.Events exposing (onClick)
import Page.Project.Commit.Route as CommitRoute
import Page.Project.Route as ProjectRoute
import Route exposing (Route)
import Util exposing ((=>))
import Views.Build exposing (viewBuildStatusIconClasses, viewBuildTextClass)
import Views.Commit exposing (branchList, infoPanel, truncateCommitMessage)
import Views.Helpers exposing (onClickPage)
import Views.Style as Style


-- CONFIG --


type alias Config msg =
    { animateMsg : Animation.Msg -> msg
    , hideCollapsableSidebarMsg : msg
    , showCollapsableSidebarMsg : msg
    , toggleSidebarMsg : msg
    , newUrlMsg : String -> msg
    , direction : Direction
    }


type DisplayType
    = Fixed FixedVisibility
    | Collapsable CollapsableVisibility Size


type FixedVisibility
    = FixedHidden
    | FixedVisible Size


type Size
    = Normal
    | ExtraWide


type Direction
    = Left
    | Right


type CollapsableVisibility
    = Visible Animation.State
    | Hidden Animation.State



-- SUBSCRIPTIONS --


subscriptions : Config msg -> DisplayType -> Sub msg
subscriptions { animateMsg } displayType =
    case displayType of
        Collapsable (Visible animationState) _ ->
            Animation.subscription animateMsg [ animationState ]

        Collapsable (Hidden animationState) _ ->
            Animation.subscription animateMsg [ animationState ]

        _ ->
            Sub.none



-- UPDATE --


show : Config msg -> DisplayType -> DisplayType
show { direction } displayType =
    let
        animateShow animationState size =
            let
                animation =
                    Animation.interrupt [ Animation.to (animationFinishAttrs direction) ] animationState
            in
            Collapsable (Visible animation) size
    in
    case displayType of
        Collapsable (Hidden animationState) size ->
            animateShow animationState size

        Collapsable (Visible animationState) size ->
            animateShow animationState size

        _ ->
            displayType


hide : Config msg -> DisplayType -> DisplayType
hide { direction } displayType =
    let
        animateHide animationState size =
            let
                animation =
                    Animation.interrupt [ Animation.to (animationStartAttrs direction size) ] animationState
            in
            Collapsable (Hidden animation) size
    in
    case displayType of
        Collapsable (Visible animationState) size ->
            animateHide animationState size

        Collapsable (Hidden animationState) size ->
            animateHide animationState size

        _ ->
            displayType


toggle : Config msg -> DisplayType -> DisplayType
toggle config displayType =
    case displayType of
        Collapsable (Visible _) _ ->
            hide config displayType

        Collapsable (Hidden _) _ ->
            show config displayType

        _ ->
            displayType


animate : Config msg -> DisplayType -> Animation.Msg -> DisplayType
animate config displayType msg =
    case displayType of
        Collapsable (Visible animationState) size ->
            let
                animation =
                    Animation.update msg animationState
            in
            Collapsable (Visible animation) size

        Collapsable (Hidden animationState) size ->
            let
                animation =
                    Animation.update msg animationState
            in
            Collapsable (Hidden animation) size

        _ ->
            displayType



-- VIEW --


view : Config msg -> DisplayType -> Html.Html msg -> Html.Html msg
view config displayType content =
    content
        |> sidebar config displayType
        |> toUnstyled


viewLogo : (String -> msg) -> Styled.Html msg
viewLogo newUrlMsg =
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


sidebarWidth : DisplayType -> Float
sidebarWidth sidebarType =
    case sidebarType of
        Collapsable _ Normal ->
            75

        Collapsable _ ExtraWide ->
            308

        Fixed (FixedVisible Normal) ->
            75

        Fixed (FixedVisible ExtraWide) ->
            308

        Fixed FixedHidden ->
            0


isCollapsable : DisplayType -> Bool
isCollapsable sidebarType =
    case sidebarType of
        Collapsable _ _ ->
            True

        _ ->
            False


fixedVisibleExtraWide : DisplayType
fixedVisibleExtraWide =
    Fixed (FixedVisible ExtraWide)


normalSize : Size
normalSize =
    Normal


extraWideSize : Size
extraWideSize =
    ExtraWide


collapsableVisible : Direction -> DisplayType
collapsableVisible direction =
    Collapsable (Visible <| Animation.style (animationFinishAttrs direction)) ExtraWide


animationStartAttrs : Direction -> Size -> List Animation.Property
animationStartAttrs direction size =
    let
        pxDistance =
            case size of
                Normal ->
                    -75.0

                ExtraWide ->
                    -308.0
    in
    case direction of
        Left ->
            animateLeft pxDistance

        Right ->
            animateRight pxDistance


animationFinishAttrs : Direction -> List Animation.Property
animationFinishAttrs direction =
    case direction of
        Left ->
            animateLeft 0.0

        Right ->
            animateRight 0.0


animateLeft : Float -> List Animation.Property
animateLeft left =
    [ Animation.left (Animation.px left) ]


animateRight : Float -> List Animation.Property
animateRight right =
    [ Animation.right (Animation.px right) ]


sidebar : Config msg -> DisplayType -> Html.Html msg -> Styled.Html msg
sidebar config displayType content =
    div
        (List.concat
            [ sidebarAnimationAttrs displayType
            , [ css
                    [ sidebarBaseStyle
                    , sidebarStyle displayType
                    ]
              ]
            ]
        )
        [ fromUnstyled content ]


collapsableOverlay : DisplayType -> List Style
collapsableOverlay displayType =
    case displayType of
        Collapsable (Visible _) _ ->
            [ position fixed
            , top (px 56)
            , right (px 0)
            , bottom (px 0)
            , zIndex (int 1)
            , width (pct 100)
            ]

        _ ->
            [ display none ]


sidebarAnimationAttrs : DisplayType -> List (Attribute msg)
sidebarAnimationAttrs displayType =
    case displayType of
        Collapsable (Visible animationState) _ ->
            animationToStyledAttrs animationState

        Collapsable (Hidden animationState) _ ->
            animationToStyledAttrs animationState

        Fixed _ ->
            []


animationToStyledAttrs : Animation.State -> List (Attribute msg)
animationToStyledAttrs animationState =
    animationState
        |> Animation.render
        |> List.map Attributes.fromUnstyled


sidebarStyle : DisplayType -> Style
sidebarStyle displayType =
    case displayType of
        Fixed (FixedVisible ExtraWide) ->
            width (px 308)

        Fixed (FixedVisible Normal) ->
            width (px 75)

        Fixed FixedHidden ->
            display none

        Collapsable _ _ ->
            width (px 308)


sidebarBaseStyle : Style
sidebarBaseStyle =
    Css.batch
        [ top (px 0)
        , bottom (px 0)
        , zIndex (int 1)
        , backgroundColor (hex "ffffff")
        , color (rgb 66 82 110)
        , position fixed
        ]


sizeWidth : Size -> Int
sizeWidth size =
    case size of
        Normal ->
            75

        ExtraWide ->
            308


initDisplayType : Config msg -> Device.Size -> Maybe DisplayType -> Size -> DisplayType
initDisplayType { direction } deviceWidth displayType size =
    if Device.isLarge deviceWidth then
        Fixed (FixedVisible size)
    else
        case displayType of
            Just (Collapsable (Visible animationState) oldSize) ->
                if oldSize == size then
                    Collapsable (Visible animationState) size
                else
                    let
                        animation =
                            Animation.interrupt [ Animation.set (animationFinishAttrs direction) ] animationState
                    in
                    Collapsable (Visible animation) size

            Just (Collapsable (Hidden animationState) oldSize) ->
                if oldSize == size then
                    Collapsable (Hidden animationState) size
                else
                    let
                        animation =
                            Animation.interrupt [ Animation.set (animationStartAttrs direction size) ] animationState
                    in
                    Collapsable (Hidden animation) size

            _ ->
                Collapsable (Hidden (Animation.style (animationStartAttrs direction size))) Normal
