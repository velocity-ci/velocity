module Data.Commit exposing (Commit, Hash(..), compare, decodeHash, decoder, hashParser, hashToString, truncateHash)

import Data.Branch as Branch exposing (Branch)
import Data.Helpers exposing (stringToDateTime)
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, optional, required)
import Time.DateTime as DateTime exposing (DateTime)
import UrlParser


type alias Commit =
    { branches : List Branch.Name
    , hash : Hash
    , author : String
    , date : DateTime
    , message : String
    }



-- SERIALIZATION --


decoder : Decoder Commit
decoder =
    decode Commit
        |> optional "branches" (Decode.map Branch.Name Decode.string |> Decode.list) []
        |> required "hash" decodeHash
        |> required "author" Decode.string
        |> required "createdAt" stringToDateTime
        |> required "message" Decode.string



-- IDENTIFIERS --


type Hash
    = Hash String


hashParser : UrlParser.Parser (Hash -> a) a
hashParser =
    UrlParser.custom "HASH" (Ok << Hash)


hashToString : Hash -> String
hashToString (Hash slug) =
    slug


decodeHash : Decoder Hash
decodeHash =
    Decode.map Hash Decode.string



-- PUBLIC HELPERS --


compare : Commit -> Commit -> Bool
compare a b =
    hashToString a.hash == hashToString b.hash


truncateHash : Hash -> String
truncateHash hash =
    hash
        |> hashToString
        |> String.slice 0 7
