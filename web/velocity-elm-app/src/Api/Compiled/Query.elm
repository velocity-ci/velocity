-- Do not manually edit this file, it was auto-generated by dillonkearns/elm-graphql
-- https://github.com/dillonkearns/elm-graphql


module Api.Compiled.Query exposing (listKnownHosts, listProjects)

import Api.Compiled.InputObject
import Api.Compiled.Interface
import Api.Compiled.Object
import Api.Compiled.Scalar
import Api.Compiled.Union
import Graphql.Internal.Builder.Argument as Argument exposing (Argument)
import Graphql.Internal.Builder.Object as Object
import Graphql.Internal.Encode as Encode exposing (Value)
import Graphql.Operation exposing (RootMutation, RootQuery, RootSubscription)
import Graphql.OptionalArgument exposing (OptionalArgument(..))
import Graphql.SelectionSet exposing (SelectionSet)
import Json.Decode as Decode exposing (Decoder)


{-| Get all known hosts
-}
listKnownHosts : SelectionSet decodesTo Api.Compiled.Object.KnownHost -> SelectionSet (Maybe (List (Maybe decodesTo))) RootQuery
listKnownHosts object_ =
    Object.selectionForCompositeField "listKnownHosts" [] object_ (identity >> Decode.nullable >> Decode.list >> Decode.nullable)


{-| List projects
-}
listProjects : SelectionSet decodesTo Api.Compiled.Object.Project -> SelectionSet (Maybe (List (Maybe decodesTo))) RootQuery
listProjects object_ =
    Object.selectionForCompositeField "listProjects" [] object_ (identity >> Decode.nullable >> Decode.list >> Decode.nullable)