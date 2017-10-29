module Data.Branch exposing (decoder, Branch, Name(..), nameParser, nameToString)

import Json.Decode as Decode exposing (Decoder)
import UrlParser


type alias Branch =
    Name



-- SERIALIZATION --


decoder : Decoder Branch
decoder =
    Decode.map Name Decode.string



-- IDENTIFIERS --


type Name
    = Name String


nameParser : UrlParser.Parser (Name -> a) a
nameParser =
    UrlParser.custom "NAME" (Ok << Name)


nameToString : Name -> String
nameToString (Name slug) =
    slug
