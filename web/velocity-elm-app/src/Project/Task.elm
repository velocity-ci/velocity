module Project.Task exposing (Task, decoder)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, optional, required)
import Json.Encode as Encode
import Project.Commit as Commit exposing (Commit)
import Project.Task.Id as Id exposing (Id)
import Project.Task.Name as Name exposing (Name)
import Project.Task.Slug as Slug exposing (Slug)
import Project.Task.Step as Step exposing (Step)


type Task
    = Task Internals


type alias Internals =
    { id : Id
    , slug : Slug
    , name : Name
    , description : String
    , steps : List Step
    , parameters : List Parameter
    , commit : Commit
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



-- Decoders


decoder : Decoder Task
decoder =
    Decode.succeed Task
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "id" Id.decoder
        |> required "slug" Slug.decoder
        |> required "name" Name.decoder
        |> optional "description" Decode.string ""
        |> optional "steps" (Decode.list Step.decoder) []
        |> optional "parameters" (Decode.list parameterDecoder) []
        |> required "commit" Commit.decoder


stringParameterDecoder : Decoder StringParameter
stringParameterDecoder =
    Decode.succeed StringParameter
        |> required "name" Decode.string
        |> optional "default" (Decode.nullable Decode.string) Nothing
        |> optional "secret" Decode.bool False


choiceParameterDecoder : Decoder ChoiceParameter
choiceParameterDecoder =
    Decode.succeed ChoiceParameter
        |> required "name" Decode.string
        |> optional "default" (Decode.nullable Decode.string) Nothing
        |> optional "secret" Decode.bool False
        |> required "otherOptions" (Decode.list Decode.string)


derivedParameterDecoder : Decoder DerivedParameter
derivedParameterDecoder =
    Decode.succeed DerivedParameter
        |> required "use" Decode.string


parameterDecoder : Decoder Parameter
parameterDecoder =
    Decode.string
        |> Decode.field "type"
        |> Decode.andThen
            (\paramType ->
                case paramType of
                    "basic" ->
                        basicParameterDecoder

                    "derived" ->
                        Decode.map DerivedParam derivedParameterDecoder

                    unknown ->
                        Decode.fail <| "Unknown parameter type: " ++ unknown
            )


basicParameterDecoder : Decoder Parameter
basicParameterDecoder =
    Decode.string
        |> Decode.list
        |> Decode.nullable
        |> Decode.field "otherOptions"
        |> Decode.andThen
            (\otherOptions ->
                let
                    string =
                        Decode.map StringParam stringParameterDecoder

                    choice =
                        Decode.map ChoiceParam choiceParameterDecoder
                in
                    case otherOptions of
                        Nothing ->
                            string

                        Just options ->
                            if List.isEmpty options then
                                string
                            else
                                choice
            )
