module Request.Commit exposing (list, get, tasks, task, builds, createBuild)

import Context exposing (Context)
import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.Project as Project exposing (Project)
import Data.Commit as Commit exposing (Commit)
import Data.Task as Task exposing (Task)
import Data.Branch as Branch exposing (Branch)
import Data.Build as Build exposing (Build)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Json.Encode as Encode
import Request.Helpers exposing (apiUrl)
import Request.Errors
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Util exposing ((=>))
import Http
import Task as ElmTask


baseUrl : String
baseUrl =
    "/projects"



-- COMMITS --


list :
    Context
    -> Project.Slug
    -> Maybe Branch.Name
    -> Int
    -> Int
    -> Maybe AuthToken
    -> ElmTask.Task Request.Errors.HttpError (PaginatedList Commit)
list context projectSlug maybeBranch amount page maybeToken =
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
        apiUrl context (baseUrl ++ "/" ++ Project.slugToString projectSlug ++ "/commits")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> HttpBuilder.withQueryParams queryParams
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> ElmTask.mapError Request.Errors.handleHttpError



-- GET --


get : Context -> Project.Slug -> Commit.Hash -> Maybe AuthToken -> ElmTask.Task Request.Errors.HttpError Commit
get context projectSlug hash maybeToken =
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
        apiUrl context (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> ElmTask.mapError Request.Errors.handleHttpError



-- TASKS --


tasks : Context -> Project.Slug -> Commit.Hash -> Maybe AuthToken -> ElmTask.Task Request.Errors.HttpError (PaginatedList Task)
tasks context projectSlug hash maybeToken =
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
            , "tasks?amount=-1"
            ]
    in
        apiUrl context (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> ElmTask.mapError Request.Errors.handleHttpError


task : Context -> Project.Slug -> Commit.Hash -> Task.Name -> Maybe AuthToken -> ElmTask.Task Request.Errors.HttpError Task
task context projectSlug hash name maybeToken =
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
        apiUrl context (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> ElmTask.mapError Request.Errors.handleHttpError



-- BUILDS --


builds : Context -> Project.Slug -> Commit.Hash -> Maybe AuthToken -> ElmTask.Task Request.Errors.HttpError (PaginatedList Build)
builds context projectSlug hash maybeToken =
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
            , "builds?amount=-1"
            ]
    in
        apiUrl context (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> ElmTask.mapError Request.Errors.handleHttpError


createBuild : Context -> Project.Slug -> Commit.Hash -> Task.Name -> List ( String, String ) -> AuthToken -> ElmTask.Task Request.Errors.HttpError Build
createBuild context projectSlug hash taskName params token =
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
        apiUrl context (String.join "/" urlPieces)
            |> HttpBuilder.post
            |> HttpBuilder.withExpect expect
            |> withAuthorization (Just token)
            |> withBody body
            |> HttpBuilder.toTask
            |> ElmTask.mapError Request.Errors.handleHttpError
