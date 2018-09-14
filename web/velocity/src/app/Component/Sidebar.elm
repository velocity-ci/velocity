module Component.Sidebar
    exposing
        ( view
        , animate
        , show
        , hide
        , toggle
        , subscriptions
        , Config
        , initDisplayType
        , fixedVisibleExtraWide
        , sidebarWidth
        , containerMarginLeft
        , sidebarAnimationAttrs
        , sidebarLeft
        , normalSize
        , extraWideSize
        , collapsableVisible
        , isCollapsable
        , DisplayType
        , collapsableOverlay
        , Size
        )

-- INTERNAL

import Data.Commit as Commit exposing (Commit)
import Data.Task as Task exposing (Task)
import Data.Project as Project exposing (Project)
import Data.Build as Build exposing (Build)
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Commit exposing (branchList, infoPanel, truncateCommitMessage)
import Views.Helpers exposing (onClickPage)
import Views.Build exposing (viewBuildStatusIconClasses, viewBuildTextClass)
import Views.Style as Style
import Util exposing ((=>))


-- EXTERNAL

import Html exposing (Html)
import Html.Styled.Attributes as Attributes exposing (css, class, classList)
import Html.Styled as Styled exposing (..)
import Html.Styled.Events exposing (onClick)
import Css exposing (..)
import Animation


-- CONFIG --


type alias Config msg =
    { animateMsg : Animation.Msg -> msg
    , hideCollapsableSidebarMsg : msg
    , showCollapsableSidebarMsg : msg
    , toggleSidebarMsg : msg
    , newUrlMsg : String -> msg
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


show : DisplayType -> DisplayType
show displayType =
    case Debug.log "DISPLAY TYPE" displayType of
        Collapsable (Hidden animationState) size ->
            let
                animation =
                    Animation.interrupt [ Animation.to animationFinishAttrs ] animationState
            in
                Collapsable (Visible animation) size

        _ ->
            displayType


hide : DisplayType -> DisplayType
hide displayType =
    case displayType of
        Collapsable (Visible animationState) size ->
            let
                animation =
                    Animation.interrupt [ Animation.to (animationStartAttrs size) ] animationState
            in
                Collapsable (Hidden animation) size

        _ ->
            displayType


toggle : DisplayType -> DisplayType
toggle displayType =
    case displayType of
        Collapsable (Visible _) _ ->
            hide displayType

        Collapsable (Hidden _) _ ->
            show displayType

        _ ->
            displayType


animate : DisplayType -> Animation.Msg -> DisplayType
animate displayType msg =
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


containerMarginLeft : DisplayType -> Float
containerMarginLeft sidebarType =
    case sidebarType of
        Collapsable _ _ ->
            0.0

        Fixed (FixedVisible Normal) ->
            75

        Fixed (FixedVisible ExtraWide) ->
            308

        Fixed FixedHidden ->
            0


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


sidebarLeft : DisplayType -> Float
sidebarLeft sidebarType =
    case sidebarType of
        Collapsable _ ExtraWide ->
            -308

        Collapsable _ Normal ->
            -75.0

        --        Collapsable (Visible _) ExtraWide ->
        --            0.0
        --
        --        Collapsable (Visible _) Normal ->
        --            0.0
        Fixed (FixedVisible Normal) ->
            0.0

        Fixed (FixedVisible ExtraWide) ->
            0.0

        Fixed FixedHidden ->
            -1000.0


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


collapsableVisible : DisplayType
collapsableVisible =
    Collapsable (Visible <| Animation.style animationFinishAttrs) ExtraWide



--collapsableHidden : DisplayType
--collapsableHidden =
--    Collapsable (Hidden <| Animation.style animationStartAttrs) ExtraWide


animationStartAttrs : Size -> List Animation.Property
animationStartAttrs size =
    case size of
        Normal ->
            animateLeft -75.0

        ExtraWide ->
            animateLeft -308.0


animationFinishAttrs : List Animation.Property
animationFinishAttrs =
    animateLeft 0.0


animateLeft : Float -> List Animation.Property
animateLeft left =
    [ Animation.left (Animation.px left) ]



--sidebarContainer : Config msg -> DisplayType -> Html.Html msg -> Styled.Html msg
--sidebarContainer config displayType content =
--    div []
--        [ div
--            [ css (collapsableOverlay displayType)
--            , onClick config.hideCollapsableSidebarMsg
--            ]
--            []
--        , sidebar config displayType content
--        ]


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
              --            , backgroundColor (hex "000000")
              --            , opacity (num 0.1)
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


initDisplayType : Int -> Maybe DisplayType -> Size -> DisplayType
initDisplayType windowWidth displayType size =
    if windowWidth >= 992 then
        Fixed (FixedVisible size)
    else
        case displayType of
            Just (Collapsable (Visible _) _) ->
                Collapsable (Visible (Animation.style animationFinishAttrs)) size

            Just (Collapsable (Hidden _) _) ->
                Collapsable (Hidden (Animation.style (animationStartAttrs size))) size

            _ ->
                Collapsable (Hidden (Animation.style (animationStartAttrs size))) Normal
