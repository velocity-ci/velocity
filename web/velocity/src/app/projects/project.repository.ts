import { Injectable } from '@angular/core';
import 'rxjs/add/operator/toPromise';
import { APIService } from "../api.service";
import { Project } from "../models/project.model";

@Injectable()
export class ProjectRepository {

    private projects: Map<string, Project>

    constructor(
        private apiService: APIService,
    ) {
        this.projects = new Map();
    }

    public refreshProjects() {
      this.apiService.get('/v1/projects')
        .then(res => {
          const projects = res.json() as Project[];
          for (const p of projects) {
            this.projects.set(p.id, p);
          }
        });
    }

    public getAllProjects(): Project[] {
        return Array.from(this.projects.values());
    }

    public getProjectById(id: string): Project {
      if (this.projects.has(id)) {
        return this.projects.get(id);
      }
      return null;
    }

    public createProject(p: Project): Promise<Project> {
        return this.apiService.post('/v1/projects', p)
            .then(res => res.json() as Project)
            .then(project => {this.projects.set(project.id, project); return project})
        ;
    }

    public updateProject(p: Project) {

    }

    public deleteProject() {

    }
}
