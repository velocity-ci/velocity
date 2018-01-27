module Data.Build exposing (..)

import Data.Task as Task
import Data.BuildStep as BuildStep exposing (BuildStep)
import Json.Decode as Decode exposing (Decoder, int, string)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required, optional)
import Data.Helpers exposing (stringToDateTime)
import Time.DateTime as DateTime exposing (DateTime)
import UrlParser


type alias Build =
    { id : Id
    , status : Status
    , taskId : Task.Id
    , steps : List BuildStep
    , createdAt : DateTime
    }



-- SERIALIZATION --


decoder : Decoder Build
decoder =
    decode Build
        |> required "id" (Decode.map Id string)
        |> required "status" statusDecoder
        |> required "task" Task.decodeId
        |> required "buildSteps" (Decode.list BuildStep.decoder)
        |> required "createdAt" stringToDateTime


statusDecoder : Decoder Status
statusDecoder =
    Decode.string
        |> Decode.andThen
            (\status ->
                case status of
                    "waiting" ->
                        Decode.succeed Waiting

                    "failed" ->
                        Decode.succeed Failed

                    "running" ->
                        Decode.succeed Running

                    "success" ->
                        Decode.succeed Success

                    unknown ->
                        Decode.fail <| "Unknown status: " ++ unknown
            )



-- IDENTIFIERS --


compare : Build -> Build -> Bool
compare a b =
    idToString a.id == idToString b.id


idParser : UrlParser.Parser (Id -> a) a
idParser =
    UrlParser.custom "ID" (Ok << Id)


idQueryParser : String -> UrlParser.QueryParser (Maybe String -> b) b
idQueryParser id =
    UrlParser.customParam id identity


type Status
    = Waiting
    | Failed
    | Running
    | Success


type Id
    = Id String


idToString : Id -> String
idToString (Id id) =
    id
