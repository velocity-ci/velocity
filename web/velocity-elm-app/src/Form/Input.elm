module Form.Input exposing (Config, text)

import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Events exposing (onClick)
import Element.Font as Font
import Element.Input as Input
import Form
import Icon
import Palette



-- Text


type alias Config msg =
    { leftIcon : Maybe (Icon.Options -> Element msg)
    , rightIcon : Maybe (Icon.Options -> Element msg)
    , label : Input.Label msg
    , placeholder : Maybe String
    , dirty : Bool
    , value : String
    , problems : List String
    , onChange : String -> msg
    }


text : Config msg -> Element msg
text config =
    let
        valid =
            List.isEmpty config.problems

        dirty =
            config.dirty
    in
    el
        (List.concat
            [ statusAttrs valid dirty
            , [ height (px 40)
              , width fill
              , Border.width 1
              , Border.rounded 4
              , Font.size 16
              , inFront (maybeIconLeft config.leftIcon)
              , inFront (maybeIconRight config.rightIcon)
              , mouseOver
                    [ Border.shadow
                        { offset = ( 0, 0 )
                        , size = 1
                        , blur = 0
                        , color = Palette.neutral5
                        }
                    ]
              ]
            ]
        )
        (Input.text
            [ Input.focusedOnLoad
            , Border.width 0
            , Background.color Palette.transparent
            , paddingXY 30 0
            , height fill
            , focused (statusDecorations valid dirty)
            ]
            { onChange = config.onChange
            , placeholder = maybePlaceholder config.placeholder
            , text = config.value
            , label = Input.labelHidden "Repository URL"
            }
        )


maybePlaceholder : Maybe String -> Maybe (Input.Placeholder msg)
maybePlaceholder =
    Maybe.map
        (\placeholderString ->
            Input.placeholder
                [ width shrink
                , height shrink
                , centerY
                , Font.color Palette.neutral4
                ]
                (Element.text placeholderString)
        )


maybeIconLeft : Maybe (Icon.Options -> Element msg) -> Element msg
maybeIconLeft maybeIcon =
    maybeIcon
        |> Maybe.map leftIcon
        |> Maybe.withDefault none


maybeIconRight : Maybe (Icon.Options -> Element msg) -> Element msg
maybeIconRight maybeIcon =
    maybeIcon
        |> Maybe.map rightIcon
        |> Maybe.withDefault none



-- Icons


iconOpts : Icon.Options
iconOpts =
    let
        opts =
            Icon.defaultOptions
    in
    { opts | size = 16 }


rightIcon : (Icon.Options -> Element msg) -> Element msg
rightIcon icon =
    el
        (List.concat
            [ iconAttrs
            , [ alignLeft
              , moveRight 7
              ]
            ]
        )
        (icon iconOpts)


leftIcon : (Icon.Options -> Element msg) -> Element msg
leftIcon icon =
    el
        (List.concat
            [ iconAttrs
            , [ alignLeft
              , moveRight 7
              ]
            ]
        )
        (icon iconOpts)


iconAttrs : List (Attribute msg)
iconAttrs =
    [ width shrink
    , height shrink
    , centerY
    ]



-- Validity Attrs


statusAttrs : Bool -> Bool -> List (Element.Attribute msg)
statusAttrs valid dirty =
    case ( valid, dirty ) of
        ( True, True ) ->
            [ Font.color Palette.success3
            , Border.color Palette.success3
            , focused (statusDecorations valid dirty)
            ]

        ( False, True ) ->
            [ Font.color Palette.danger4
            , Border.color Palette.danger6
            , focused (statusDecorations valid dirty)
            ]

        ( _, False ) ->
            [ Font.color Palette.neutral3
            , Border.color Palette.neutral3
            , focused (statusDecorations valid dirty)
            ]


statusDecorations : Bool -> Bool -> List Decoration
statusDecorations valid dirty =
    []



--    if valid then
--        [ Font.color Palette.success1
--        , Border.color Palette.success5
--        , Border.shadow { offset = ( 0, 0 ), size = 2, blur = 2, color = Palette.success7 }
--        ]
--
--    else if dirty then
--        [ Font.color Palette.neutral2
--        , Border.color Palette.neutral2
--        , Border.shadow { offset = ( 0, 0 ), size = 2, blur = 2, color = Palette.neutral6 }
--        ]
--
--    else
--        [ Font.color Palette.neutral2
--        , Border.color Palette.neutral2
--        , Border.shadow { offset = ( 0, 0 ), size = 2, blur = 2, color = Palette.neutral6 }
--        ]
