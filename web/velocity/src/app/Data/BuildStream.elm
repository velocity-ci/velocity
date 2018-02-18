module Data.BuildStream exposing (..)

import Data.Project as Project
import Data.Commit as Commit
import Data.Task as Task
import Json.Decode as Decode exposing (Decoder, int, string)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required, optional)
import UrlParser
import Data.Helpers exposing (stringToDateTime)
import Time.DateTime as DateTime exposing (DateTime)


type alias BuildStream =
    { id : Id
    , name : String
    }


type alias BuildStreamOutput =
    { line : Int
    , timestamp : DateTime
    , output : String
    }



-- SERIALIZATION --


decoder : Decoder BuildStream
decoder =
    decode BuildStream
        |> required "id" (Decode.map Id string)
        |> required "name" Decode.string


outputDecoder : Decoder BuildStreamOutput
outputDecoder =
    decode BuildStreamOutput
        |> required "lineNumber" Decode.int
        |> required "timestamp" stringToDateTime
        |> required "output" Decode.string



-- IDENTIFIERS --


idParser : UrlParser.Parser (Id -> a) a
idParser =
    UrlParser.custom "ID" (Ok << Id)


type Id
    = Id String


idToString : Id -> String
idToString (Id id) =
    id
