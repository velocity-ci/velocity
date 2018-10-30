module Data.BuildStream exposing (BuildStream, BuildStreamOutput, Id(..), decoder, idParser, idToString, outputDecoder)

import Data.Commit as Commit
import Data.Helpers exposing (stringToDateTime)
import Data.Project as Project
import Data.Task as Task
import Json.Decode as Decode exposing (Decoder, int, string)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, optional, required)
import Time.DateTime as DateTime exposing (DateTime)
import UrlParser


type alias BuildStream =
    { id : Id
    , name : String
    }


type alias BuildStreamOutput =
    { line : Int
    , timestamp : DateTime
    , rawTimestamp : String
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
        |> required "timestamp" Decode.string
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
