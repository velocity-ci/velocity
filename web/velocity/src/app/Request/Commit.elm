module Request.Commit exposing (list, get, tasks, task, builds, createBuild)

import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.Project as Project exposing (Project)
import Data.Commit as Commit exposing (Commit)
import Data.Task as Task exposing (Task)
import Data.Branch as Branch exposing (Branch)
import Data.Build as Build exposing (Build)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Json.Encode as Encode
import Request.Helpers exposing (apiUrl)
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Util exposing ((=>))
import Http


baseUrl : String
baseUrl =
    "/projects"



-- COMMITS --


list :
    Project.Slug
    -> Maybe Branch.Name
    -> Int
    -> Int
    -> Maybe AuthToken
    -> Http.Request (PaginatedList Commit)
list projectSlug maybeBranch amount page maybeToken =
    let
        expect =
            Commit.decoder
                |> PaginatedList.decoder
                |> Http.expectJson

        branchParam queryParams =
            case maybeBranch of
                Just branch ->
                    ( "branch", Branch.nameToString (Just branch) ) :: queryParams

                _ ->
                    queryParams

        amountParam queryParams =
            ( "amount", toString amount ) :: queryParams

        pageParam queryParams =
            ( "page", toString page ) :: queryParams

        queryParams =
            []
                |> branchParam
                |> amountParam
                |> pageParam
    in
        apiUrl (baseUrl ++ "/" ++ Project.slugToString projectSlug ++ "/commits")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> HttpBuilder.withQueryParams queryParams
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- GET --


get : Project.Slug -> Commit.Hash -> Maybe AuthToken -> Http.Request Commit
get projectSlug hash maybeToken =
    let
        expect =
            Commit.decoder
                |> Http.expectJson

        urlPieces =
            [ baseUrl
            , Project.slugToString projectSlug
            , "commits"
            , Commit.hashToString hash
            ]
    in
        apiUrl (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- TASKS --


tasks : Project.Slug -> Commit.Hash -> Maybe AuthToken -> Http.Request (PaginatedList Task)
tasks projectSlug hash maybeToken =
    let
        expect =
            Task.decoder
                |> PaginatedList.decoder
                |> Http.expectJson

        urlPieces =
            [ baseUrl
            , Project.slugToString projectSlug
            , "commits"
            , Commit.hashToString hash
            , "tasks"
            ]
    in
        apiUrl (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest


task : Project.Slug -> Commit.Hash -> Task.Name -> Maybe AuthToken -> Http.Request Task
task projectSlug hash name maybeToken =
    let
        expect =
            Task.decoder
                |> Http.expectJson

        urlPieces =
            [ baseUrl
            , Project.slugToString projectSlug
            , "commits"
            , Commit.hashToString hash
            , "tasks"
            , Task.nameToString name
            ]
    in
        apiUrl (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- BUILDS --


builds : Project.Slug -> Commit.Hash -> Maybe AuthToken -> Http.Request (PaginatedList Build)
builds projectSlug hash maybeToken =
    let
        expect =
            Build.decoder
                |> PaginatedList.decoder
                |> Http.expectJson

        urlPieces =
            [ baseUrl
            , Project.slugToString projectSlug
            , "commits"
            , Commit.hashToString hash
            , "builds"
            ]
    in
        apiUrl (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest


createBuild : Project.Slug -> Commit.Hash -> Task.Name -> List ( String, String ) -> AuthToken -> Http.Request Build
createBuild projectSlug hash taskName params token =
    let
        expect =
            Build.decoder
                |> Http.expectJson

        urlPieces =
            [ baseUrl
            , Project.slugToString projectSlug
            , "commits"
            , Commit.hashToString hash
            , "tasks"
            , Task.nameToString taskName
            , "builds"
            ]

        encodedParams =
            let
                enc =
                    params
                        |> List.map
                            (\( field, value ) ->
                                Encode.object
                                    [ ( "name", Encode.string field )
                                    , ( "value", Encode.string value )
                                    ]
                            )
            in
                if (List.length params) > 0 then
                    Encode.list enc
                else
                    Encode.list []

        encodedBody =
            Encode.object
                [ "params" => encodedParams ]

        body =
            encodedBody
                |> Http.jsonBody
    in
        apiUrl (String.join "/" urlPieces)
            |> HttpBuilder.post
            |> HttpBuilder.withExpect expect
            |> withAuthorization (Just token)
            |> withBody body
            |> HttpBuilder.toRequest
