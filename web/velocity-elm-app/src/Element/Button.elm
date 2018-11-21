module Element.Button exposing (ButtonConfig, Scheme(..), Size(..), button, link, simpleButton, simpleLink)

import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input as Input
import Icon
import Palette
import Route exposing (Route)


type Scheme
    = Primary
    | Secondary
    | Transparent


type Size
    = Small
    | Medium
    | Large


type alias ButtonConfig msg =
    { leftIcon : Maybe (Icon.Options -> Element msg)
    , rightIcon : Maybe (Icon.Options -> Element msg)
    , centerLeftIcon : Maybe (Icon.Options -> Element msg)
    , centerRightIcon : Maybe (Icon.Options -> Element msg)
    , scheme : Scheme
    , content : Element msg
    , size : Size
    , widthLength : Length
    , disabled : Bool
    }


simpleButton : msg -> { content : Element msg, scheme : Scheme } -> Element msg
simpleButton onClick { content, scheme } =
    button onClick
        { leftIcon = Nothing
        , rightIcon = Nothing
        , centerLeftIcon = Nothing
        , centerRightIcon = Nothing
        , content = content
        , scheme = scheme
        , size = Medium
        , widthLength = fill
        , disabled = False
        }


simpleLink : Route -> { content : Element msg, scheme : Scheme } -> Element msg
simpleLink route { content, scheme } =
    link route
        { leftIcon = Nothing
        , rightIcon = Nothing
        , centerLeftIcon = Nothing
        , centerRightIcon = Nothing
        , content = content
        , scheme = scheme
        , size = Medium
        , widthLength = fill
        , disabled = False
        }


buttonIconOptions : Size -> Icon.Options
buttonIconOptions fromSize =
    let
        opts =
            Icon.defaultOptions

        toSize =
            sizeFloat fromSize
    in
    { opts | size = toSize }


link : Route -> ButtonConfig msg -> Element msg
link route ({ size, widthLength, scheme, centerLeftIcon, centerRightIcon } as buttonConfig) =
    let
        content =
            row [ centerX ]
                [ el [] (viewMaybeIcon size centerLeftIcon)
                , contentEl buttonConfig
                , el [] (viewMaybeIcon size centerRightIcon)
                ]
    in
    Route.link
        (List.concat
            [ baseAttrs widthLength
            , sizeAttrs size
            , sideIcons buttonConfig
            , schemeAttrs scheme
            , activeAttrs buttonConfig
            ]
        )
        content
        route


button : msg -> ButtonConfig msg -> Element msg
button onClickMsg ({ size, scheme, centerLeftIcon, centerRightIcon, widthLength } as buttonConfig) =
    row
        (List.concat
            [ baseAttrs widthLength
            , sideIcons buttonConfig
            , schemeAttrs scheme
            , activeAttrs buttonConfig
            , sizeAttrs size
            , [ onClick onClickMsg ]
            ]
        )
        [ el [ centerX ] (viewMaybeIcon size centerLeftIcon)
        , contentEl buttonConfig
        , el [ centerX ] (viewMaybeIcon size centerRightIcon)
        ]


contentEl : ButtonConfig msg -> Element msg
contentEl buttonConfig =
    el [ centerX, width shrink ] buttonConfig.content


sideIcons : ButtonConfig msg -> List (Attribute msg)
sideIcons { size, rightIcon, leftIcon } =
    [ inFront (el [ alignRight, centerY, moveLeft 10 ] (viewMaybeIcon size rightIcon))
    , inFront (el [ alignLeft, centerY, moveRight 10 ] (viewMaybeIcon size leftIcon))
    ]


viewMaybeIcon : Size -> Maybe (Icon.Options -> Element msg) -> Element msg
viewMaybeIcon size maybeIcon =
    case maybeIcon of
        Just icon ->
            icon (buttonIconOptions size)

        Nothing ->
            none


sizeFloat : Size -> Float
sizeFloat =
    sizeInt >> toFloat


sizeInt : Size -> Int
sizeInt fromSize =
    case fromSize of
        Small ->
            13

        Medium ->
            16

        Large ->
            21


baseAttrs : Length -> List (Element.Attribute msg)
baseAttrs widthLength =
    [ width widthLength
    , height (px 45)
    , Border.rounded 5
    , padding 11
    , alignBottom
    , spacingXY 5 0
    ]


sizeAttrs : Size -> List (Element.Attribute msg)
sizeAttrs fromSize =
    [ Font.size (sizeInt fromSize)
    ]


activeAttrs : ButtonConfig msg -> List (Attribute msg)
activeAttrs { disabled, scheme } =
    if disabled then
        [ alpha 0.6 ]

    else
        [ mouseOver (schemeMouseOverDecorations scheme)
        , pointer
        ]


schemeMouseOverDecorations : Scheme -> List Element.Decoration
schemeMouseOverDecorations scheme =
    case scheme of
        Primary ->
            [ Background.color Palette.primary4
            , Font.color Palette.white
            ]

        Secondary ->
            [ Background.color Palette.neutral5
            , Font.color Palette.primary2
            ]

        Transparent ->
            [ Background.color Palette.transparent
            , Font.color Palette.primary2
            ]


schemeAttrs : Scheme -> List (Element.Attribute msg)
schemeAttrs scheme =
    case scheme of
        Primary ->
            [ Border.color Palette.primary4
            , Font.color Palette.white
            , Background.color Palette.primary3
            , Border.width 1
            ]

        Secondary ->
            [ Border.color Palette.neutral6
            , Font.color Palette.primary3
            , Background.color Palette.neutral6
            , Border.width 1
            ]

        Transparent ->
            [ Border.color Palette.neutral4
            , Background.color Palette.transparent
            , Font.color Palette.primary3
            , Border.width 0
            ]
