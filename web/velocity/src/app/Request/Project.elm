module Request.Project
    exposing
        ( list
        , create
        , get
        , commits
        , commit
        , commitTasks
        , commitTask
        , sync
        , delete
        , branches
        )

import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.Project as Project exposing (Project)
import Data.Commit as Commit exposing (Commit)
import Data.Task as Task exposing (Task)
import Data.Branch as Branch exposing (Branch)
import Data.CommitResults as CommitResults
import Json.Decode as Decode
import Json.Encode as Encode
import Request.Helpers exposing (apiUrl)
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Util exposing ((=>))
import Http


baseUrl : String
baseUrl =
    "/projects"



-- LIST --


list : Maybe AuthToken -> Http.Request (List Project)
list maybeToken =
    let
        expect =
            Project.decoder
                |> Decode.list
                |> Http.expectJson
    in
        apiUrl baseUrl
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- SYNC --


sync : Project.Id -> AuthToken -> Http.Request Project
sync id token =
    let
        expect =
            Project.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.idToString id ++ "/sync")
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withExpect expect
            |> HttpBuilder.toRequest



-- BRANCHES --


branches : Project.Id -> Maybe AuthToken -> Http.Request (List Branch)
branches id maybeToken =
    let
        expect =
            Decode.list Branch.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.idToString id ++ "/branches")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- COMMITS --


commits : Project.Id -> Maybe Branch -> Maybe AuthToken -> Http.Request CommitResults.Results
commits id maybeBranch maybeToken =
    let
        expect =
            CommitResults.decoder
                |> Http.expectJson

        queryParams =
            case maybeBranch of
                Just (Branch.Name branch) ->
                    [ ( "branch", branch ) ]

                _ ->
                    []
    in
        apiUrl (baseUrl ++ "/" ++ Project.idToString id ++ "/commits")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> HttpBuilder.withQueryParams queryParams
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest


commit : Project.Id -> Commit.Hash -> Maybe AuthToken -> Http.Request Commit
commit id hash maybeToken =
    let
        expect =
            Commit.decoder
                |> Http.expectJson

        urlPieces =
            [ baseUrl
            , Project.idToString id
            , "commits"
            , Commit.hashToString hash
            ]
    in
        apiUrl (String.join "/" urlPieces)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest


commitTasks : Project.Id -> Commit.Hash -> Maybe AuthToken -> Http.Request (List Task)
commitTasks id hash maybeToken =
    let
        expect =
            Task.decoder
                |> Decode.list
                |> Http.expectJson

        urlPieces =
            [ baseUrl
            , Project.idToString id
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


commitTask : Project.Id -> Commit.Hash -> Task.Name -> Maybe AuthToken -> Http.Request Task
commitTask id hash name maybeToken =
    let
        expect =
            Task.decoder
                |> Http.expectJson

        urlPieces =
            [ baseUrl
            , Project.idToString id
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



-- GET --


get : Project.Id -> Maybe AuthToken -> Http.Request Project
get id maybeToken =
    let
        expect =
            Project.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.idToString id)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- CREATE --


type alias CreateConfig record =
    { record
        | name : String
        , repository : String
        , privateKey : Maybe String
    }


create : CreateConfig record -> AuthToken -> Http.Request Project
create config token =
    let
        expect =
            Project.decoder
                |> Http.expectJson

        baseProject =
            [ "name" => Encode.string config.name
            , "repository" => Encode.string config.repository
            ]

        project =
            case config.privateKey of
                Just privateKey ->
                    ( "key", Encode.string privateKey ) :: baseProject

                Nothing ->
                    baseProject

        body =
            project
                |> Encode.object
                |> Http.jsonBody
    in
        apiUrl baseUrl
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withBody body
            |> withExpect expect
            |> HttpBuilder.toRequest



-- DELETE --


delete : Project.Id -> AuthToken -> Http.Request ()
delete id token =
    apiUrl (baseUrl ++ "/" ++ Project.idToString id)
        |> HttpBuilder.delete
        |> withAuthorization (Just token)
        |> withExpect (Http.expectStringResponse (\_ -> Ok ()))
        |> HttpBuilder.toRequest
