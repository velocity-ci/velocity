-- Do not manually edit this file, it was auto-generated by dillonkearns/elm-graphql
-- https://github.com/dillonkearns/elm-graphql


module Api.Compiled.Subscription exposing (knownHostAdded, knownHostVerified, projectAdded)

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
import Json.Decode as Decode exposing (Decoder)


knownHostAdded : SelectionSet decodesTo Api.Compiled.Object.KnownHost -> SelectionSet decodesTo RootSubscription
knownHostAdded object_ =
    Object.selectionForCompositeField "knownHostAdded" [] object_ identity


knownHostVerified : SelectionSet decodesTo Api.Compiled.Object.KnownHost -> SelectionSet decodesTo RootSubscription
knownHostVerified object_ =
    Object.selectionForCompositeField "knownHostVerified" [] object_ identity


projectAdded : SelectionSet decodesTo Api.Compiled.Object.Project -> SelectionSet decodesTo RootSubscription
projectAdded object_ =
    Object.selectionForCompositeField "projectAdded" [] object_ identity
