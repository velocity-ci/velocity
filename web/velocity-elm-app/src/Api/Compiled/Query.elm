-- Do not manually edit this file, it was auto-generated by dillonkearns/elm-graphql
-- https://github.com/dillonkearns/elm-graphql


module Api.Compiled.Query exposing (forHost, knownHosts, projects, users)

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


{-| Get fingerprint for host
-}
forHost : SelectionSet decodesTo Api.Compiled.Object.KnownHost -> SelectionSet (Maybe decodesTo) RootQuery
forHost object_ =
    Object.selectionForCompositeField "forHost" [] object_ (identity >> Decode.nullable)


{-| Get all known hosts
-}
knownHosts : SelectionSet decodesTo Api.Compiled.Object.KnownHost -> SelectionSet (Maybe (List (Maybe decodesTo))) RootQuery
knownHosts object_ =
    Object.selectionForCompositeField "knownHosts" [] object_ (identity >> Decode.nullable >> Decode.list >> Decode.nullable)


{-| Get all projects
-}
projects : SelectionSet decodesTo Api.Compiled.Object.Project -> SelectionSet (Maybe (List (Maybe decodesTo))) RootQuery
projects object_ =
    Object.selectionForCompositeField "projects" [] object_ (identity >> Decode.nullable >> Decode.list >> Decode.nullable)


{-| Get all users
-}
users : SelectionSet decodesTo Api.Compiled.Object.User -> SelectionSet (Maybe (List (Maybe decodesTo))) RootQuery
users object_ =
    Object.selectionForCompositeField "users" [] object_ (identity >> Decode.nullable >> Decode.list >> Decode.nullable)
