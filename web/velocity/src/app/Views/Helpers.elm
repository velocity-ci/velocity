module Views.Helpers exposing (onClickPage, styledOnClickPage)

import Html.Events exposing (onWithOptions, defaultOptions)
import Html.Styled.Attributes as StyledAttribute
import Html.Styled
import Html exposing (Attribute)
import Json.Decode exposing (Decoder)
import Route exposing (Route)


styledOnClickPage : (String -> msg) -> Route -> Html.Styled.Attribute msg
styledOnClickPage msg route =
    onClickPage msg route
        |> StyledAttribute.fromUnstyled


onClickPage : (String -> msg) -> Route -> Attribute msg
onClickPage msg route =
    onPreventDefaultClick (msg (Route.routeToString route))


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
