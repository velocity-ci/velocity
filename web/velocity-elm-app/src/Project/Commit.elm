module Project.Commit exposing (Commit, decoder, hash)

import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, optional, required)
import Project.Branch.Name as BranchName
import Project.Commit.Hash as Hash exposing (Hash)
import Time


type Commit
    = Commit Internals


type alias Internals =
    { branches : List BranchName.Name
    , hash : Hash
    , author : String
    , date : Time.Posix
    , message : String
    }



-- SERIALIZATION --


decoder : Decoder Commit
decoder =
    Decode.succeed Commit
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> optional "branches" (Decode.list BranchName.decoder) []
        |> required "hash" Hash.decoder
        |> required "author" Decode.string
        |> required "createdAt" Iso8601.decoder
        |> required "message" Decode.string



-- INFO


hash : Commit -> Hash
hash (Commit commit) =
    commit.hash
