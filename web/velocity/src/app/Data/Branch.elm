module Data.Branch exposing (decoder, Branch, Name(..), nameParser, nameToString, selectDecoder, allBranchName)

import Html.Events exposing (targetValue)
import Json.Decode as Decode exposing (Decoder)
import UrlParser
import Http


type alias Branch =
    Name



-- SERIALIZATION --


decoder : Decoder Branch
decoder =
    Decode.map Name (Decode.at [ "name" ] Decode.string)


selectDecoder : Decoder (Maybe Branch)
selectDecoder =
    targetValue
        |> Decode.andThen
            (\branchName ->
                if branchName == allBranchName then
                    Decode.succeed Nothing
                else
                    Name branchName
                        |> Just
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
