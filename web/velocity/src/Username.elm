module Username exposing (Username, decoder, encode, fromString, selectionSet, toHtml, toString, urlParser)

import Api.Compiled.Object
import Api.Compiled.Object.User as User
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import Html exposing (Html)
import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode exposing (Value)
import Url.Parser



-- TYPES


type Username
    = Username String


fromString : String -> Username
fromString =
    Username



-- CREATE


selectionSet : SelectionSet Username Api.Compiled.Object.User
selectionSet =
    SelectionSet.map Username User.username


decoder : Decoder Username
decoder =
    Decode.map Username Decode.string



-- TRANSFORM


encode : Username -> Value
encode (Username username) =
    Encode.string username


toString : Username -> String
toString (Username username) =
    username


urlParser : Url.Parser.Parser (Username -> a) a
urlParser =
    Url.Parser.custom "USERNAME" (\str -> Just (Username str))


toHtml : Username -> Html msg
toHtml (Username username) =
    Html.text username
