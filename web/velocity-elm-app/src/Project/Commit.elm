module Project.Commit exposing (Commit, decoder, hash, head)

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint
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


head : Cred -> BaseUrl -> ProjectSlug.Slug -> BranchName.Name -> Task Http.Error (Maybe Commit)
head cred baseUrl projectSlug branchName =
    let
        l =
            list cred baseUrl projectSlug branchName { amount = 1, page = 1 }
    in
        Http.toTask l
            |> Task.andThen (PaginatedList.values >> List.head >> Task.succeed)



-- COLLECTION --


list : Cred -> BaseUrl -> ProjectSlug.Slug -> BranchName.Name -> Endpoint.CollectionOptions -> Http.Request (PaginatedList Commit)
list cred baseUrl projectSlug branchName opts =
    let
        endpoint =
            Endpoint.commits (Just opts) (Api.toEndpoint baseUrl) projectSlug branchName
    in
        PaginatedList.decoder decoder
            |> Api.get endpoint (Just cred)
