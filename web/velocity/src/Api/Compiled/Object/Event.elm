-- Do not manually edit this file, it was auto-generated by dillonkearns/elm-graphql
-- https://github.com/dillonkearns/elm-graphql


module Api.Compiled.Object.Event exposing (id, knownHost, project, type_, user)

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


{-| The ID of an object
-}
id : SelectionSet Api.Compiled.ScalarCodecs.Id Api.Compiled.Object.Event
id =
    Object.selectionForField "ScalarCodecs.Id" "id" [] (Api.Compiled.ScalarCodecs.codecs |> Api.Compiled.Scalar.unwrapCodecs |> .codecId |> .decoder)


knownHost : SelectionSet decodesTo Api.Compiled.Object.KnownHost -> SelectionSet (Maybe decodesTo) Api.Compiled.Object.Event
knownHost object_ =
    Object.selectionForCompositeField "knownHost" [] object_ (identity >> Decode.nullable)


project : SelectionSet decodesTo Api.Compiled.Object.Project -> SelectionSet (Maybe decodesTo) Api.Compiled.Object.Event
project object_ =
    Object.selectionForCompositeField "project" [] object_ (identity >> Decode.nullable)


type_ : SelectionSet String Api.Compiled.Object.Event
type_ =
    Object.selectionForField "String" "type" [] Decode.string


user : SelectionSet decodesTo Api.Compiled.Object.User -> SelectionSet decodesTo Api.Compiled.Object.Event
user object_ =
    Object.selectionForCompositeField "user" [] object_ identity
