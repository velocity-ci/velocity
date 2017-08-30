import { Component, OnInit } from '@angular/core';
import { Project } from "../models/project.model";
import {ActivatedRoute, Router} from '@angular/router';
import { Validators, FormControl, FormGroup } from "@angular/forms";
import { ProjectRepository } from "./project.repository";

@Component({
    selector: 'app-project',
    templateUrl: './project.component.html'
})
export class ProjectComponent implements OnInit {

    project = new Project();

    constructor(
      private route: ActivatedRoute,
      private projectRepository: ProjectRepository,
      private router: Router
    ) {}

    ngOnInit(): void {
      const id = this.route.snapshot.params['id'];
      if (this.projectRepository.getProjectById(id)) {
        this.project = this.projectRepository.getProjectById(id);
      }

      if (this.project.id == null) {
        this.router.navigate(['']);
      }
    }

    submit() {
        this.projectRepository.createProject(this.project).then(res => {
        });
    }
}
