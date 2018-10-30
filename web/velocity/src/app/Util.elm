module Util exposing ((=>), appendErrors, capitalize, onClickStopPropagation, pair, viewIf, viewIfStyled)

import Char
import Html exposing (Attribute, Html)
import Html.Events exposing (defaultOptions, onWithOptions)
import Html.Styled
import Json.Decode as Decode


(=>) : a -> b -> ( a, b )
(=>) =
    \a b -> ( a, b )


{-| infixl 0 means the (=>) operator has the same precedence as (<|) and (|>),
meaning you can use it at the end of a pipeline and have the precedence work out.
-}
infixl 0 =>


{-| Useful when building up a Cmd via a pipeline, and then pairing it with
a model at the end.

    session.user
        |> User.Request.foo
        |> Task.attempt Foo
        |> pair { model | something = blah }

-}
pair : a -> b -> ( a, b )
pair first second =
    first => second


viewIf : Bool -> Html msg -> Html msg
viewIf condition content =
    if condition then
        content
    else
        Html.text ""


viewIfStyled : Bool -> Html.Styled.Html msg -> Html.Styled.Html msg
viewIfStyled condition content =
    if condition then
        content
    else
        Html.Styled.text ""


onClickStopPropagation : msg -> Attribute msg
onClickStopPropagation msg =
    onWithOptions "click"
        { defaultOptions | stopPropagation = True }
        (Decode.succeed msg)


appendErrors : { model | errors : List error } -> List error -> { model | errors : List error }
appendErrors model errors =
    { model | errors = model.errors ++ errors }


capitalize : String -> String
capitalize string =
    string
        |> String.uncons
        |> Maybe.map
            (\( a, b ) ->
                let
                    upper =
                        Char.toUpper a
                in
                    String.cons upper b
            )
        |> Maybe.withDefault string
