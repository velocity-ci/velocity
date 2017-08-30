import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import {NgbModule} from '@ng-bootstrap/ng-bootstrap';

import { AppComponent } from './app.component';
import { RouterModule } from '@angular/router';

import { IndexComponent } from './index/index.component';
import { SignInComponent } from "./authentication/sign-in.component";
import { APIService } from "./api.service";
import { AuthService } from "./authentication/authentication.service";
import { HttpModule } from "@angular/http";
import { ProjectsComponent } from "./projects/projects.component";
import { ProjectRepository } from "./projects/project.repository";
import { ProjectComponent } from "./projects/project.component";


@NgModule({
  declarations: [
    AppComponent,
    IndexComponent,
    SignInComponent,
    ProjectsComponent,
    ProjectComponent
  ],
  imports: [
    BrowserModule,
    NgbModule.forRoot(),
    FormsModule,
    ReactiveFormsModule,
    HttpModule,
    RouterModule.forRoot([
      { path: '', component: IndexComponent},
      { path: 'sign-in', component: SignInComponent},
      { path: 'projects', component: ProjectsComponent, canActivate: [AuthService]},
      { path: 'projects/:id', component: ProjectComponent, canActivate: [AuthService]},
    ], { useHash: true })
  ],
  providers: [
    AuthService,
    APIService,
    ProjectRepository,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
