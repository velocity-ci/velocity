module Data.Project exposing (Project, decoder, idParser, idToString, decodeId, Id(..))

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required, optional)
import Time.DateTime as DateTime exposing (DateTime)
import Data.Helpers exposing (stringToDateTime)
import UrlParser


type alias Project =
    { id : Id
    , name : String
    , repository : String
    , createdAt : DateTime
    , updatedAt : DateTime
    , synchronising : Bool
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    decode Project
        |> required "id" decodeId
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" stringToDateTime
        |> required "updatedAt" stringToDateTime
        |> required "synchronising" Decode.bool



-- IDENTIFIERS --


type Id
    = Id String


decodeId : Decoder Id
decodeId =
    Decode.map Id Decode.string


idParser : UrlParser.Parser (Id -> a) a
idParser =
    UrlParser.custom "ID" (Ok << Id)


idToString : Id -> String
idToString (Id slug) =
    slug
