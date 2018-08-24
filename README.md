[//]: # " DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS HEADER. "
[//]: # "  "
[//]: # " Copyright (c) 2018 Oracle and/or its affiliates. All rights reserved. "
[//]: # "  "
[//]: # " The contents of this file are subject to the terms of either the GNU "
[//]: # " General Public License Version 2 only (''GPL'') or the Common Development "
[//]: # " and Distribution License(''CDDL'') (collectively, the ''License'').  You "
[//]: # " may not use this file except in compliance with the License.  You can "
[//]: # " obtain a copy of the License at "
[//]: # " https://oss.oracle.com/licenses/CDDL+GPL-1.1 "
[//]: # " or LICENSE.txt.  See the License for the specific "
[//]: # " language governing permissions and limitations under the License. "
[//]: # "  "
[//]: # " When distributing the software, include this License Header Notice in each "
[//]: # " file and include the License file at LICENSE.txt. "
[//]: # "  "
[//]: # " GPL Classpath Exception: "
[//]: # " Oracle designates this particular file as subject to the ''Classpath'' "
[//]: # " exception as provided by Oracle in the GPL Version 2 section of the License "
[//]: # " file that accompanied this code. "
[//]: # "  "
[//]: # " Modifications: "
[//]: # " If applicable, add the following below the License Header, with the fields "
[//]: # " enclosed by brackets [] replaced by your own identifying information: "
[//]: # " ''Portions Copyright [year] [name of copyright owner]'' "
[//]: # "  "
[//]: # " Contributor(s): "
[//]: # " If you wish your version of this file to be governed by only the CDDL or "
[//]: # " only the GPL Version 2, indicate your decision by adding ''[Contributor] "
[//]: # " elects to include this software in this distribution under the [CDDL or GPL "
[//]: # " Version 2] license.''  If you don't indicate a single choice of license, a "
[//]: # " recipient has the option to distribute your version of this file under "
[//]: # " either the CDDL, the GPL Version 2 or to extend the choice of license to "
[//]: # " its licensees as provided above.  However, if you add GPL Version 2 code "
[//]: # " and therefore, elected the GPL Version 2 license, then the option applies "
[//]: # " only if the new code is made subject to such option by the copyright "
[//]: # " holder. "

# dargo

Dynamic Service Registry and Inversion of Control for GO

## Service Registry

Dargo is an in-memory service registry for GO.  It also introduces inversion of control, in that once the
service descriptions are bound into the registry they are created in response to registry lookups.  Services
are scoped by context, and so are created based on the lifecycle defined by the scope.  For
example a service in the Singleton scope are only every created once.  A service in the PerLookup
scope are created every time they are looked up.

NOTE:  The current version of this API is 0.1.0.  This means that the API has
not settled completely and may change in future revisions.  Once the dargo
team has decided the API is good as it is we will make the 1.0 version which
will have some backward compatibility guarantees.  In the meantime, if you
have questions or comments please open issues.  Thank you.

The general flow of an application that uses dargo is to:

1.  Create a ServiceLocator
2.  Bind services into the ServiceLocator
3.  Use the ServiceLocator in your code to find services

Services can depend on other services.  When a service is created first all of its dependencies can be created with
the same ServiceLocator.

There can be multiple implementations of the same service, and there are specific rules
for choosing the best service amongst all of the possible choices.  In some cases services can be differentiated
by qualifiers.  In other cases services can be given ranks, with higher ranks being chosen over lower ranks.

Using dargo helps unit test your code as it becomes easy to replace services served by the locator with mocks.
If you ensure that your test mocks have a higher rank than the service bound by your normal code then
all of your internal code will use the mock from the ServiceLocator rather than the original service.

### An Example

In the following example a MusicService depends on seven NoteServices.  Each NoteService is qualified with
one of the letters from A to G.


