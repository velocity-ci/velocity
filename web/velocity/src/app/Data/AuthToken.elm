module Data.AuthToken exposing (AuthToken, encode, decoder, withAuthorization, tokenToString)

import Json.Encode as Encode exposing (Value)
import Json.Decode as Decode exposing (Decoder)
import HttpBuilder exposing (withHeader, RequestBuilder)


type AuthToken
    = AuthToken String



-- SERIALIZATION --


encode : AuthToken -> Value
encode (AuthToken token) =
    Encode.string token


decoder : Decoder AuthToken
decoder =
    Decode.string
        |> Decode.map AuthToken



-- IDENTIFIERS --


tokenToString : AuthToken -> String
tokenToString (AuthToken token) =
    token



-- HELPERS --


withAuthorization : Maybe AuthToken -> RequestBuilder a -> RequestBuilder a
withAuthorization maybeToken builder =
    case maybeToken of
        Just (AuthToken token) ->
            builder
                |> withHeader "Authorization" ("Bearer " ++ token)

        Nothing ->
            builder
