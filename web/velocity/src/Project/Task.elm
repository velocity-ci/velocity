module Project.Task exposing (Task, byBranch, description, name, selectionSet)

import Api exposing (BaseUrl, Cred)
import Api.Compiled.Object
import Api.Compiled.Object.Task as Task
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)
import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, optional, required)
import Project.Branch.Name as BranchName
import Project.Commit as Commit exposing (Commit)
import Project.Slug as ProjectSlug
import Project.Task.Id as Id exposing (Id)
import Project.Task.Name as Name exposing (Name)
import Project.Task.Slug as Slug exposing (Slug)
import Project.Task.Step as Step exposing (Step)
import Task as BaseTask


type Task
    = Task Internals


type alias Internals =
    { --    id : Id
      --    , slug : Slug
      name : Name
    , description : Maybe String

    --    , steps : List Step
    --    , parameters : List Parameter
    --    , commit : Commit
    }


type Parameter
    = StringParam StringParameter
    | ChoiceParam ChoiceParameter
    | DerivedParam DerivedParameter


type alias StringParameter =
    { name : String
    , default : Maybe String
    , secret : Bool
    }


type alias ChoiceParameter =
    { name : String
    , default : Maybe String
    , secret : Bool
    , options : List String
    }


type alias DerivedParameter =
    { use : String }



-- Info


name : Task -> Name
name (Task t) =
    t.name


description : Task -> String
description (Task t) =
    Maybe.withDefault "" t.description



-- Decoders


selectionSet : SelectionSet Task Api.Compiled.Object.Task
selectionSet =
    SelectionSet.map Task internalSelectionSet


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.Task
internalSelectionSet =
    SelectionSet.succeed Internals
        |> with Name.selectionSet
        |> with Task.description



--
--stringParameterDecoder : Decoder StringParameter
--stringParameterDecoder =
--    Decode.succeed StringParameter
--        |> required "name" Decode.string
--        |> optional "default" (Decode.nullable Decode.string) Nothing
--        |> optional "secret" Decode.bool False
--
--
--choiceParameterDecoder : Decoder ChoiceParameter
--choiceParameterDecoder =
--    Decode.succeed ChoiceParameter
--        |> required "name" Decode.string
--        |> optional "default" (Decode.nullable Decode.string) Nothing
--        |> optional "secret" Decode.bool False
--        |> required "otherOptions" (Decode.list Decode.string)
--
--
--derivedParameterDecoder : Decoder DerivedParameter
--derivedParameterDecoder =
--    Decode.succeed DerivedParameter
--        |> required "use" Decode.string
--
--
--parameterDecoder : Decoder Parameter
--parameterDecoder =
--    Decode.string
--        |> Decode.field "type"
--        |> Decode.andThen
--            (\paramType ->
--                case paramType of
--                    "basic" ->
--                        basicParameterDecoder
--
--                    "derived" ->
--                        Decode.map DerivedParam derivedParameterDecoder
--
--                    unknown ->
--                        Decode.fail <| "Unknown parameter type: " ++ unknown
--            )
--
--
--basicParameterDecoder : Decoder Parameter
--basicParameterDecoder =
--    Decode.string
--        |> Decode.list
--        |> Decode.nullable
--        |> Decode.field "otherOptions"
--        |> Decode.andThen
--            (\otherOptions ->
--                let
--                    string =
--                        Decode.map StringParam stringParameterDecoder
--
--                    choice =
--                        Decode.map ChoiceParam choiceParameterDecoder
--                in
--                    case otherOptions of
--                        Nothing ->
--                            string
--
--                        Just options ->
--                            if List.isEmpty options then
--                                string
--                            else
--                                choice
--            )
-- COLLECTION


byBranch : Cred -> BaseUrl -> ProjectSlug.Slug -> BranchName.Name -> BaseTask.Task Http.Error (List Task)
byBranch cred baseUrl projectSlug branchName =
    BaseTask.succeed []



--    let
--        endpoint =
--            Endpoint.tasks (Just { amount = -1, page = 1 }) (Api.toEndpoint baseUrl) projectSlug
--    in
--    Commit.head cred baseUrl projectSlug branchName
--        |> BaseTask.andThen
--            (\maybeCommit ->
--                case maybeCommit of
--                    Just commit ->
--                        PaginatedList.decoder decoder
--                            |> Api.get (endpoint <| Commit.hash commit) (Just cred)
--                            |> Cmd.map PaginatedList.values
--
--                    Nothing ->
--                        BaseTask.succeed []
--            )
