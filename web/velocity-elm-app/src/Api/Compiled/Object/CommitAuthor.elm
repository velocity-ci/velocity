-- Do not manually edit this file, it was auto-generated by dillonkearns/elm-graphql
-- https://github.com/dillonkearns/elm-graphql


module Api.Compiled.Object.CommitAuthor exposing (date, email, id, name)

import Api.Compiled.InputObject
import Api.Compiled.Interface
import Api.Compiled.Object
import Api.Compiled.Scalar
import Api.Compiled.ScalarCodecs
import Api.Compiled.Union
import Graphql.Internal.Builder.Argument as Argument exposing (Argument)
import Graphql.Internal.Builder.Object as Object
import Graphql.Internal.Encode as Encode exposing (Value)
import Graphql.Operation exposing (RootMutation, RootQuery, RootSubscription)
import Graphql.OptionalArgument exposing (OptionalArgument(..))
import Graphql.SelectionSet exposing (SelectionSet)
import Json.Decode as Decode


date : SelectionSet Api.Compiled.ScalarCodecs.NaiveDateTime Api.Compiled.Object.CommitAuthor
date =
    Object.selectionForField "ScalarCodecs.NaiveDateTime" "date" [] (Api.Compiled.ScalarCodecs.codecs |> Api.Compiled.Scalar.unwrapCodecs |> .codecNaiveDateTime |> .decoder)


email : SelectionSet String Api.Compiled.Object.CommitAuthor
email =
    Object.selectionForField "String" "email" [] Decode.string


{-| The ID of an object
-}
id : SelectionSet Api.Compiled.ScalarCodecs.Id Api.Compiled.Object.CommitAuthor
id =
    Object.selectionForField "ScalarCodecs.Id" "id" [] (Api.Compiled.ScalarCodecs.codecs |> Api.Compiled.Scalar.unwrapCodecs |> .codecId |> .decoder)


name : SelectionSet String Api.Compiled.Object.CommitAuthor
name =
    Object.selectionForField "String" "name" [] Decode.string
