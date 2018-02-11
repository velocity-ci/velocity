module Data.User exposing (User, Username, decoder, encode, usernameToString, usernameParser, usernameToHtml, usernameDecoder)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, required)
import Json.Encode as Encode exposing (Value)
import Data.AuthToken as AuthToken exposing (AuthToken)
import UrlParser
import Util exposing ((=>))
import Html exposing (Html)


type alias User =
    { username : Username
    , token : AuthToken
    }



-- SERIALIZATION --


decoder : Decoder User
decoder =
    decode User
        |> required "username" usernameDecoder
        |> required "token" AuthToken.decoder


encode : User -> Value
encode user =
    Encode.object
        [ "username" => encodeUsername user.username
        , "token" => AuthToken.encode user.token
        ]



-- IDENTIFIERS --


type Username
    = Username String


usernameToString : Username -> String
usernameToString (Username username) =
    username


usernameParser : UrlParser.Parser (Username -> a) a
usernameParser =
    UrlParser.custom "USERNAME" (Ok << Username)


usernameDecoder : Decoder Username
usernameDecoder =
    Decode.map Username Decode.string


encodeUsername : Username -> Value
encodeUsername (Username username) =
    Encode.string username


usernameToHtml : Username -> Html msg
usernameToHtml (Username username) =
    Html.text username
