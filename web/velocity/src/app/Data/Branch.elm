module Data.Branch exposing (decoder, Branch, Name(..), nameParser, nameToString, selectDecoder, allBranchName)

import Html.Events exposing (targetValue)
import Json.Decode as Decode exposing (Decoder)
import UrlParser
import Http
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required, optional)


type alias Branch =
    { name : Name
    , active : Bool
    }



-- SERIALIZATION --


decoder : Decoder Branch
decoder =
    decode Branch
        |> required "name" (Decode.map Name Decode.string)
        |> required "active" Decode.bool


selectDecoder : Decoder (Maybe Branch)
selectDecoder =
    targetValue
        |> Decode.andThen
            (\branchName ->
                if branchName == allBranchName then
                    Decode.succeed Nothing
                else
                    Just { name = Name branchName, active = True }
                        |> Decode.succeed
            )



-- IDENTIFIERS --


type Name
    = Name String


allBranchName : String
allBranchName =
    "all-branches"


nameParser : UrlParser.Parser (Maybe Name -> a) a
nameParser =
    UrlParser.custom "NAME" <|
        (\s ->
            let
                maybeBranch =
                    if s == allBranchName then
                        Nothing
                    else
                        s
                            |> Http.decodeUri
                            |> Maybe.map Name
            in
                Ok maybeBranch
        )


nameToString : Maybe Name -> String
nameToString maybeName =
    case maybeName of
        Just (Name slug) ->
            slug

        Nothing ->
            allBranchName
