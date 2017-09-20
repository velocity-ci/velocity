module Data.Project exposing (Project, decoder, idParser, idToString, Id)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required)
import Time.DateTime as DateTime exposing (DateTime)
import Data.Helpers exposing (stringToDateTime)
import UrlParser


type alias Project =
    { id : Id
    , key : String
    , name : String
    , repository : String
    , createdAt : DateTime
    , updatedAt : DateTime
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    decode Project
        |> required "id" (Decode.map Id Decode.string)
        |> required "key" Decode.string
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" stringToDateTime
        |> required "updatedAt" stringToDateTime



-- IDENTIFIERS --


type Id
    = Id String


idParser : UrlParser.Parser (Id -> a) a
idParser =
    UrlParser.custom "ID" (Ok << Id)


idToString : Id -> String
idToString (Id slug) =
    slug
