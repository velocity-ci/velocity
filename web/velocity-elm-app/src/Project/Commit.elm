module Project.Commit exposing (Commit, hash)

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



-- INFO


hash : Commit -> Hash
hash (Commit commit) =
    commit.hash
