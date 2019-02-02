-- Do not manually edit this file, it was auto-generated by dillonkearns/elm-graphql
-- https://github.com/dillonkearns/elm-graphql


module Api.Compiled.Object.BranchConnection exposing (edges, pageInfo)

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


edges : SelectionSet decodesTo Api.Compiled.Object.BranchEdge -> SelectionSet (Maybe (List (Maybe decodesTo))) Api.Compiled.Object.BranchConnection
edges object_ =
    Object.selectionForCompositeField "edges" [] object_ (identity >> Decode.nullable >> Decode.list >> Decode.nullable)


pageInfo : SelectionSet decodesTo Api.Compiled.Object.PageInfo -> SelectionSet decodesTo Api.Compiled.Object.BranchConnection
pageInfo object_ =
    Object.selectionForCompositeField "pageInfo" [] object_ identity