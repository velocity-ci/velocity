module Views.Helpers exposing (onPreventDefaultClick)

import Html.Attributes exposing (href)
import Html.Events exposing (onWithOptions, defaultOptions)
import Html exposing (Attribute)
import Json.Decode exposing (Decoder)


--onClickPage : Attribute msg -> String -> List (Attribute msg)
--onClickPage msg url =
--    [ href url
--    , onPreventDefaultClick msg
--    ]


onPreventDefaultClick : msg -> Attribute msg
onPreventDefaultClick message =
    onWithOptions "click"
        { defaultOptions | preventDefault = True }
        (preventDefault2
            |> Json.Decode.andThen (maybePreventDefault message)
        )


preventDefault2 : Decoder Bool
preventDefault2 =
    Json.Decode.map2
        (invertedOr)
        (Json.Decode.field "ctrlKey" Json.Decode.bool)
        (Json.Decode.field "metaKey" Json.Decode.bool)


maybePreventDefault : msg -> Bool -> Decoder msg
maybePreventDefault msg preventDefault =
    case preventDefault of
        True ->
            Json.Decode.succeed msg

        False ->
            Json.Decode.fail "Normal link"


invertedOr : Bool -> Bool -> Bool
invertedOr x y =
    not (x || y)
