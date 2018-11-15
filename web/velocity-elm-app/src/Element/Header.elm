module Element.Header exposing (notificationsToggle, userMenuToggle)

import Asset
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input as Input
import Icon
import Palette


iconOptions : Icon.Options
iconOptions =
    Icon.defaultOptions


notificationsToggle : { amount : Int, toggled : Bool, toggleMsg : Bool -> msg } -> Element msg
notificationsToggle { amount, toggled, toggleMsg } =
    el
        [ height (px 45)
        , Border.widthEach
            { top = 0
            , left = 0
            , right = 0
            , bottom =
                if toggled then
                    3

                else
                    0
            }
        , Border.color Palette.transparent
        , paddingXY 8 0
        , pointer
        , Font.color Palette.white
        , onClick (toggleMsg <| not toggled)
        , Border.color Palette.primary7
        , Font.color Palette.primary7
        ]
        (el
            [ width (px 28)
            , height fill
            , moveDown 3
            , Border.rounded 180
            , centerY
            , above
                (el
                    [ width shrink
                    , height fill
                    , Background.color Palette.primary5
                    , paddingXY 4 3
                    , Border.rounded 7
                    , moveDown 14
                    , moveRight 14
                    , Font.size 10
                    ]
                    (text "2")
                )
            ]
            (Icon.bell { iconOptions | size = 100, sizeUnit = Icon.Percentage })
        )


userMenuToggle : Element msg
userMenuToggle =
    column
        [ Background.color Palette.neutral7
        , Border.color Palette.neutral4
        , Border.width 1
        , Border.rounded 7
        , moveRight -170
        , width (px 200)
        ]
        [ row
            [ width fill
            , padding 10
            , spacingXY 10 0
            ]
            [ el
                [ width (px 45)
                , height (px 45)
                , Border.rounded 90
                , Background.image (Asset.src Asset.defaultAvatar)
                ]
                (text "")
            , column [ width fill ]
                [ paragraph [ Font.size 15, Font.color Palette.primary2 ] [ text "Signed in as" ]
                , paragraph [ Font.size 18, Font.heavy, Font.color Palette.primary5 ] [ text "admin" ]
                ]
            ]
        , row
            [ Border.widthEach { top = 1, left = 0, right = 0, bottom = 0 }
            , Border.color Palette.neutral6
            , mouseOver
                [ Background.color Palette.neutral2
                , Font.color Palette.white
                ]
            , width fill
            , paddingXY 20 20
            , spacingXY 10 0
            , Font.color Palette.primary1
            , Font.light
            , Font.size 16
            ]
            [ column
                [ width shrink ]
                [ Icon.logOut { iconOptions | size = 16 } ]
            , column
                [ width fill ]
                [ text "Sign out" ]
            ]
        ]
