module Project.Commit exposing (Commit, decoder, hash)

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint exposing (Endpoint)
import Http
import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, optional, required)
import Json.Encode as Encode
import PaginatedList exposing (PaginatedList)
import Project.Branch.Name as BranchName
import Project.Commit.Hash as Hash exposing (Hash)
import Project.Slug as ProjectSlug
import Task exposing (Task)
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



-- SINGLE --
--head : Cred -> BaseUrl -> ProjectSlug.Slug -> BranchName.Name -> (Result Http.Error a -> msg) -> Cmd msg
--head cred baseUrl projectSlug branchName toMsg =
--    let
--        endpoint =
--            commitsEndpoint cred baseUrl projectSlug branchName { amount = 1, page = 1 }
--    in
--    PaginatedList.decoder decoder
--        |> Api.get endpoint (Just cred) toMsg
--
--        |> Task.andThen (PaginatedList.values >> List.head >> Task.succeed)
-- COLLECTION --


list : Cred -> BaseUrl -> ProjectSlug.Slug -> BranchName.Name -> Endpoint.CollectionOptions -> (Result Http.Error (PaginatedList Commit) -> msg) -> Cmd msg
list cred baseUrl projectSlug branchName opts toMsg =
    let
        endpoint =
            commitsEndpoint cred baseUrl projectSlug branchName opts
    in
    Api.get endpoint (Just cred) toMsg (PaginatedList.decoder decoder)


commitsEndpoint : Cred -> BaseUrl -> ProjectSlug.Slug -> BranchName.Name -> Endpoint.CollectionOptions -> Endpoint
commitsEndpoint cred baseUrl projectSlug branchName opts =
    Endpoint.commits (Just opts) (Api.toEndpoint baseUrl) projectSlug branchName
